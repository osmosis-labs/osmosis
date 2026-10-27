"""Offline smoke tests for the release-script environment."""

import importlib.util
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import threading
import unittest

import idna

ROOT = Path(__file__).resolve().parent


def load_script(name, path):
    spec = importlib.util.spec_from_file_location(name, ROOT / path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


binaries = load_script("binaries", "create_binaries_json/create_binaries_json.py")
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
        if self.path != "/checksums":
            self.send_error(404)
            return
        self.send_response(200)
        self.end_headers()
        self.wfile.write(CHECKSUMS.encode())

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
        text = binaries.download_checksums(self.url + "/redirect")
        self.assertEqual(text, CHECKSUMS)
        result = json.loads(binaries.checksums_to_binaries_json(text))
        self.assertEqual(result, {"binaries": {
            "linux/amd64": (
                "https://github.com/osmosis-labs/osmosis/releases/download/"
                f"v32.0.0/osmosisd-32.0.0-linux-amd64?checksum=sha256:{CHECKSUM}"
            )
        }})

    def test_download_failure(self):
        with self.assertRaisesRegex(ValueError, "Status code: 404"):
            binaries.download_checksums(self.url + "/missing")

    def test_idna(self):
        self.assertEqual(idna.encode("bücher.de"), b"xn--bcher-kva.de")


if __name__ == "__main__":
    unittest.main()
