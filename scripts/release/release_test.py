"""Offline smoke tests for the shared release-script environment."""

import importlib.util
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import sys
import threading
import unittest

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


if __name__ == "__main__":
    unittest.main()
