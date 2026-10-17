#!/usr/bin/env python3
"""Deploy an immutable image while preserving the single PostgreSQL volume."""
import argparse
import json
from pathlib import Path
import subprocess
import time
import urllib.request

ROOT = Path(__file__).resolve().parents[1]


def api(method, path, body=None):
    result = subprocess.run(["bunny", "api", method, path],
                            input=json.dumps(body) if body is not None else "",
                            capture_output=True, text=True)
    if result.returncode:
        raise RuntimeError(f"Bunny request failed: {method} {path}")
    return json.loads(result.stdout) if result.stdout.strip() else None


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def validate(app, config):
    require(app["autoScaling"] == {"min": 1, "max": 1}, "Expected exactly one replica")
    for key in ("allowedRegionIds", "requiredRegionIds"):
        require(app["regionSettings"][key] == ["DE"], "Expected DE only")
    containers = {c["name"]: c for c in app["containerTemplates"]}
    require(set(containers) == {"web", "postgres"}, "Unexpected containers")
    require(containers["postgres"]["volumeMounts"] ==
            [{"name": "postgres-data", "mountPath": "/var/lib/postgresql/data"}],
            "Postgres volume mount changed")
    require(app["volumes"] == [{"name": "postgres-data", "size": 1}], "Volume configuration changed")
    require((containers["web"]["imageNamespace"], containers["web"]["imageName"]) ==
            ("klausbreyer", "pokerhans"), "Unexpected web image")
    return containers["web"]


def volume_identity(base, config):
    volumes = api("GET", base + "/volumes")["items"]
    require(len(volumes) == 1 and volumes[0]["name"] == "postgres-data", "Unexpected volume")
    instances = volumes[0]["volumeInstances"]
    require(len(instances) == 1, "Expected one volume instance")
    instance = instances[0]
    require(instance["id"] == config["volume_instance_id"], "Persistent volume was replaced")
    require(instance["status"] == "attached" and len(instance["attachedPods"]) == 1,
            "Volume must be attached to exactly one pod")
    return instance["id"]


def ready_pod(base):
    overview = api("GET", base + "/overview")
    pods = [p for r in overview["regions"] for p in r["pods"]]
    require(len([p for p in pods if p["status"] == "ready"]) <= 1, "Multiple ready pods")
    if overview["status"] == "active" and len(pods) == 1 and pods[0]["status"] == "ready":
        return pods[0]["name"]


def check_http(config):
    for path in ("/health", "/", "/seasons"):
        for attempt in range(10):
            try:
                request = urllib.request.Request(config["url"] + path, headers={"Cache-Control": "no-cache"})
                with urllib.request.urlopen(request, timeout=15) as response:
                    body = response.read()
                    require(response.status == 200, "HTTP check failed")
                    require(body == b"ok" if path == "/health" else b"Pokerhans" in body,
                            "Unexpected response")
                break
            except (OSError, RuntimeError):
                if attempt == 9:
                    raise
                time.sleep(3)


def deploy(config, tag=None):
    base = "/mc/apps/" + config["app_id"]
    web = validate(api("GET", base), config)
    old_pod = ready_pod(base)
    require(old_pod, "App must be healthy before deployment")
    volume_identity(base, config)
    if tag:
        image = api("POST", "/mc/registries/digest", {
            "registryId": web["imageRegistryId"], "imageNamespace": "klausbreyer",
            "imageName": "pokerhans", "tag": tag})
        if web["imageDigest"] != image["digest"]:
            api("PATCH", base + "/containers/" + web["id"], {
                "imageTag": tag, "imageDigest": image["digest"], "imagePullPolicy": "always"})
            print("Deploying " + tag, flush=True)
            deadline = time.monotonic() + 900
            while time.monotonic() < deadline:
                pod = ready_pod(base)
                if pod and pod != old_pod:
                    current = validate(api("GET", base), config)
                    require(current["imageDigest"] == image["digest"], "Unexpected deployed image")
                    volume_identity(base, config)
                    break
                time.sleep(5)
            else:
                raise RuntimeError("Deployment timed out. Inspect Bunny before retrying")
    check_http(config)
    print("Verified one healthy pod and the original Postgres volume", flush=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--tag", help="Immutable Git commit image tag; omit for read-only verification")
    args = parser.parse_args()
    deploy(json.loads((ROOT / "bunny.json").read_text()), args.tag)
