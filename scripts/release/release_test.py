"""Offline smoke tests for the shared release-script environment."""

from contextlib import redirect_stderr, redirect_stdout
from io import StringIO
import importlib.util
import json
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import threading
import unittest
from unittest.mock import patch

import idna

ROOT = Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT / "update_chain_registry"))


def load_script(name, path):
    spec = importlib.util.spec_from_file_location(name, ROOT / path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


binaries = load_script("binaries", "create_binaries_json/create_binaries_json.py")
registry = load_script("registry", "update_chain_registry/update_chain_registry.py")
CHECKSUM = "a" * 64
CHECKSUMS = (
    f"{CHECKSUM}  osmosisd-32.0.0-linux-amd64\n"
    f"{CHECKSUM}  osmosisd-32.0.0-linux-amd64.tar.gz\n"
)


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/redirect":
            self.send_response(302)
            self.send_header("Location", "/checksums")
            self.end_headers()
            return
        body = CHECKSUMS if self.path == "/checksums" else '{"chain_name":"osmosis"}'
        self.send_response(200)
        self.end_headers()
        self.wfile.write(body.encode())

    def log_message(self, *args):
        pass


class TestReleaseScripts(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        cls.thread = threading.Thread(target=cls.server.serve_forever, daemon=True)
        cls.thread.start()
        cls.url = f"http://127.0.0.1:{cls.server.server_port}"

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()
        cls.server.server_close()
        cls.thread.join()

    def test_checksum_download_and_parsing(self):
        for module in (binaries, registry):
            with self.subTest(script=module.__name__):
                text = module.download_checksums(self.url + "/redirect")
                self.assertEqual(text, CHECKSUMS)
                result = module.checksums_to_binaries_json(text)
                if isinstance(result, str):
                    result = json.loads(result)
                self.assertEqual(result, {"binaries": {
                    "linux/amd64": (
                        "https://github.com/osmosis-labs/osmosis/releases/download/"
                        f"v32.0.0/osmosisd-32.0.0-linux-amd64?checksum=sha256:{CHECKSUM}"
                    )
                }})

    def test_registry_fetch(self):
        self.assertEqual(registry.fetch_data(self.url + "/data", "json"),
                         {"chain_name": "osmosis"})
        self.assertEqual(registry.fetch_data(self.url + "/checksums", "text"),
                         CHECKSUMS)

    def test_idna(self):
        self.assertEqual(idna.encode("bücher.de"), b"xn--bcher-kva.de")


class TestRegistryCLI(unittest.TestCase):
    def run_cli(self, chain_json=None, error=None, version="v32.0.0"):
        stdout, stderr = StringIO(), StringIO()
        args = ["update_chain_registry.py", "--upgrade_version", version,
                "--upgrade_height", "123", "--debug"]
        with patch.object(sys, "argv", args), \
                patch.object(registry, "fetch_data", return_value=chain_json, side_effect=error), \
                patch.object(registry, "DEBUG", False), \
                redirect_stdout(stdout), redirect_stderr(stderr):
            try:
                registry.main()
                code = 0
            except SystemExit as exc:
                code = exc.code
        return code, stdout.getvalue(), stderr.getvalue()

    def test_fetch_failure_exits_nonzero_without_stdout(self):
        code, stdout, stderr = self.run_cli(error=RuntimeError("404 Not Found"))
        self.assertEqual(code, 1)
        self.assertEqual(stdout, "")
        self.assertIn("404 Not Found", stderr)

    def test_skipped_updates_exit_nonzero_without_stdout(self):
        for recommended in ("v32.0.0", "v33.0.0"):
            with self.subTest(recommended=recommended):
                code, stdout, stderr = self.run_cli(chain_json={"codebase": {
                    "recommended_version": recommended,
                    "compatible_versions": [recommended],
                }})
                self.assertEqual(code, 1)
                self.assertEqual(stdout, "")
                self.assertIn("Couldn't update codebase", stderr)

    def test_invalid_tag_exits_nonzero_without_stdout(self):
        code, stdout, stderr = self.run_cli(version="invalid")
        self.assertEqual(code, 1)
        self.assertEqual(stdout, "")
        self.assertIn("vX.Y.Z", stderr)

    def test_successful_debug_output_is_valid_json(self):
        chain_json = {"chain_name": "osmosis", "codebase": {
            "recommended_version": "v31.0.0",
            "compatible_versions": ["v31.0.0"],
            "versions": [{"name": "v31"}],
        }}
        info = {
            "name": "v32", "recommended_version": "v32.0.0",
            "compatible_versions": ["v32.0.0"], "binaries": {},
            "cosmos_sdk_version": "v0.50.15", "consensus": {},
            "cosmwasm_version": "v0.54.10", "ibc_go_version": "v8.8.0",
        }
        with patch.object(registry, "create_version_info", return_value=info):
            code, stdout, stderr = self.run_cli(chain_json=chain_json)
        self.assertEqual(code, 0)
        self.assertEqual(json.loads(stdout)["codebase"]["recommended_version"], "v32.0.0")
        self.assertIn("Major release", stderr)


class TestRegistryWrapper(unittest.TestCase):
    def test_output_is_replaced_only_after_valid_json(self):
        cases = (
            (42, "setup failed", False),
            (0, "An error occurred: 404 Not Found", False),
            (0, "", False),
            (0, '{"chain_name":"osmosis"}', True),
        )
        for exit_code, output, success in cases:
            with self.subTest(exit_code=exit_code, output=output), \
                    tempfile.TemporaryDirectory(prefix="registry wrapper ") as tmp:
                root = Path(tmp)
                script_dir = root / "scripts/release/update_chain_registry"
                script_dir.mkdir(parents=True)
                wrapper = script_dir / "run.sh"
                shutil.copy(ROOT / "update_chain_registry/run.sh", wrapper)
                target = root / "chain.schema.json"
                target.write_text("original registry\n")
                bin_dir = root / "bin"
                bin_dir.mkdir()
                fake_uv = bin_dir / "uv"
                # Model environment/script failure, while executing the real JSON validator.
                fake_uv.write_text(
                    f"#!{sys.executable}\n"
                    "import os, sys\n"
                    "args = sys.argv[1:]\n"
                    "assert args[0:2] == ['run', '--project']\n"
                    "assert args[3:5] == ['--locked', 'python']\n"
                    "assert os.path.isdir(args[2])\n"
                    "if args[5] == '-m':\n"
                    "    os.execv(sys.executable, [sys.executable] + args[5:])\n"
                    "sys.stdout.write(os.environ['FAKE_OUTPUT'])\n"
                    "sys.exit(int(os.environ['FAKE_EXIT']))\n"
                )
                fake_uv.chmod(0o755)
                env = dict(os.environ, PATH=str(bin_dir) + os.pathsep + os.environ["PATH"],
                           FAKE_OUTPUT=output, FAKE_EXIT=str(exit_code))
                result = subprocess.run(
                    ["bash", str(wrapper), "v32.0.0", "123"],
                    cwd=bin_dir, env=env, capture_output=True, text=True,
                )
                self.assertEqual(result.returncode == 0, success, result.stderr)
                self.assertEqual(target.read_text(), output if success else "original registry\n")
                self.assertEqual(list(root.glob("chain.schema.json.*")), [])


if __name__ == "__main__":
    unittest.main()
