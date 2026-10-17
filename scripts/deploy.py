#!/usr/bin/env python3
"""Deploy an immutable image using the migrated external PostgreSQL database."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import time
import urllib.request
from urllib.parse import urlsplit

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
    require(set(containers) == {"web"},
            "External database migration must finish before deployment: remove the Postgres sidecar after verification")
    require(not containers["web"].get("volumeMounts"), "Web container must not mount database storage")
    require((containers["web"]["imageNamespace"], containers["web"]["imageName"]) ==
            ("klausbreyer", "pokerhans"), "Unexpected web image")
    return containers["web"]


def database_environment(web, url):
    require(bool(url and url.strip()), "DATABASE_URL is required for deployment")
    try:
        parsed = urlsplit(url)
        valid = (parsed.scheme in ("postgres", "postgresql") and parsed.hostname
                 and parsed.hostname not in ("localhost", "127.0.0.1", "::1")
                 and parsed.path not in ("", "/"))
    except ValueError:
        valid = False
    require(valid, "DATABASE_URL must identify an external PostgreSQL database")
    environment = web.get("environmentVariables", [])
    current = next((e["value"] for e in environment if e["name"] == "DATABASE_URL"), None)
    require(current == url,
            "DATABASE_URL differs from Bunny: migrate and verify the database before changing the deployment target")
    return [e for e in environment if e["name"] != "DATABASE_URL"] + [
        {"name": "DATABASE_URL", "value": url}]


def ready_pod(base):
    overview = api("GET", base + "/overview")
    pods = [p for r in overview["regions"] for p in r["pods"]]
    # Rolling updates can briefly keep both old and new web pods ready.
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
    if tag:
        environment = database_environment(web, os.environ.get("DATABASE_URL"))
        image = api("POST", "/mc/registries/digest", {
            "registryId": web["imageRegistryId"], "imageNamespace": "klausbreyer",
            "imageName": "pokerhans", "tag": tag})
        if web["imageDigest"] != image["digest"]:
            api("PATCH", base + "/containers/" + web["id"], {
                "imageTag": tag, "imageDigest": image["digest"], "imagePullPolicy": "always",
                "environmentVariables": environment})
            print("Deploying " + tag, flush=True)
            deadline = time.monotonic() + 900
            while time.monotonic() < deadline:
                pod = ready_pod(base)
                if pod and pod != old_pod:
                    current = validate(api("GET", base), config)
                    require(current["imageDigest"] == image["digest"], "Unexpected deployed image")
                    database_environment(current, os.environ.get("DATABASE_URL"))
                    break
                time.sleep(5)
            else:
                raise RuntimeError("Deployment timed out. Inspect Bunny before retrying")
    check_http(config)
    print("Verified one healthy web pod using external PostgreSQL", flush=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--tag", help="Immutable Git commit image tag; omit for read-only verification")
    args = parser.parse_args()
    deploy(json.loads((ROOT / "bunny.json").read_text()), args.tag)
