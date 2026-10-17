"""Guard against deployments that would replace or split the production database."""
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
            "volumes": [{"name": "postgres-data", "size": 1}],
            "containerTemplates": [
                {"name": "postgres", "volumeMounts": [
                    {"name": "postgres-data", "mountPath": "/var/lib/postgresql/data"}]},
                {"name": "web", "imageNamespace": "klausbreyer", "imageName": "pokerhans"},
            ],
        }

    def test_refuses_scaling_and_changed_storage(self):
        self.assertEqual(deploy.validate(self.app, {})["name"], "web")
        for key, value in [("autoScaling", {"min": 1, "max": 2}), ("volumes", [])]:
            changed = copy.deepcopy(self.app)
            changed[key] = value
            with self.assertRaises(RuntimeError):
                deploy.validate(changed, {})

    def test_refuses_a_replacement_volume(self):
        response = {"items": [{"name": "postgres-data", "volumeInstances": [{
            "id": "new-empty-volume", "status": "attached", "attachedPods": ["pod"]}]}]}
        with patch.object(deploy, "api", return_value=response):
            with self.assertRaisesRegex(RuntimeError, "replaced"):
                deploy.volume_identity("/mc/apps/test", {"volume_instance_id": "original"})

    def test_refuses_multiple_ready_pods(self):
        response = {"status": "active", "regions": [{"pods": [
            {"name": "one", "status": "ready"}, {"name": "two", "status": "ready"}]}]}
        with patch.object(deploy, "api", return_value=response):
            with self.assertRaisesRegex(RuntimeError, "Multiple"):
                deploy.ready_pod("/mc/apps/test")


if __name__ == "__main__":
    unittest.main()
