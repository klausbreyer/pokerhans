"""Guard against deployments before a verified external database migration."""
import copy
import importlib.util
from pathlib import Path
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location("deploy", Path(__file__).parents[1] / "scripts/deploy.py")
deploy = importlib.util.module_from_spec(spec)
spec.loader.exec_module(deploy)


class DeploymentTest(unittest.TestCase):
    def setUp(self):
        self.app = {
            "autoScaling": {"min": 1, "max": 1},
            "regionSettings": {"allowedRegionIds": ["DE"], "requiredRegionIds": ["DE"]},
            "containerTemplates": [
                {"name": "web", "imageNamespace": "klausbreyer", "imageName": "pokerhans"},
            ],
        }

    def test_refuses_scaling_and_unmigrated_sidecar(self):
        self.assertEqual(deploy.validate(self.app, {})["name"], "web")
        changed = copy.deepcopy(self.app)
        changed["autoScaling"] = {"min": 1, "max": 2}
        with self.assertRaises(RuntimeError):
            deploy.validate(changed, {})
        self.app["containerTemplates"].append({"name": "postgres"})
        with self.assertRaisesRegex(RuntimeError, "migration must finish"):
            deploy.validate(self.app, {})

    def test_database_secret_preserves_other_environment(self):
        url = "postgresql://user:password@db.example.com/pokerhans?ssl=true"
        environment = [{"name": "SECRET_KEY_BASE", "value": "keep"},
                       {"name": "DATABASE_URL", "value": url}]
        self.assertEqual(deploy.database_environment({"environmentVariables": environment}, url),
                         environment)

    def test_refuses_missing_local_or_changed_database(self):
        web = {"environmentVariables": [{"name": "DATABASE_URL",
                                        "value": "postgresql://db.example.com/original"}]}
        for url in (None, "", "postgresql://localhost/pokerhans", "https://db.example.com/db",
                    "postgresql://db.example.com/another", "postgresql://[invalid/db"):
            with self.subTest(url=url), self.assertRaises(RuntimeError):
                deploy.database_environment(web, url)

    def test_waits_for_rolling_update_to_settle(self):
        response = {"status": "active", "regions": [{"pods": [
            {"name": "one", "status": "ready"}, {"name": "two", "status": "ready"}]}]}
        with patch.object(deploy, "api", return_value=response):
            self.assertIsNone(deploy.ready_pod("/mc/apps/test"))
        response["regions"][0]["pods"] = [{"name": "new", "status": "ready"}]
        with patch.object(deploy, "api", return_value=response):
            self.assertEqual(deploy.ready_pod("/mc/apps/test"), "new")


if __name__ == "__main__":
    unittest.main()
