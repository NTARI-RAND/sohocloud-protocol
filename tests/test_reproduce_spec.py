"""Exercise message selection through the independent reproducer CLI."""

import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest


FIXTURES = Path(__file__).resolve().parents[1] / "testdata"


def reproduce(*args, fixtures=FIXTURES):
    return subprocess.run(
        [sys.executable, str(fixtures / "reproduce_spec.py"), *args],
        capture_output=True,
        encoding="utf-8",
        timeout=15,
    )


class MessageSelectionTests(unittest.TestCase):
    def test_checks_only_the_selected_message(self):
        cases = (
            ("FeeDeclaration", "sig verifies over derived bytes"),
            ("OperatorRotation", "both 2-of-7 sigs verify over derived bytes"),
        )
        for name, signature_result in cases:
            with self.subTest(name=name):
                result = reproduce("--message", name)
                self.assertEqual(result.returncode, 0, result.stderr)
                checks = [line for line in result.stdout.splitlines() if line.startswith("[")]
                self.assertEqual(checks, [f"[OK] {name}", f"[OK] {name} {signature_result}"])
                self.assertIn("TOTAL CHECKS: 2  RESULT: ALL MATCH", result.stdout)

    def test_unknown_message_exits_nonzero_without_running_the_suite(self):
        result = reproduce("--message", "UnknownMessage")
        self.assertEqual(result.returncode, 2)
        self.assertIn("invalid choice", result.stderr)
        self.assertIn("UnknownMessage", result.stderr)
        self.assertEqual(result.stdout, "")

    def test_a_mismatch_fails_only_when_its_message_is_selected(self):
        with tempfile.TemporaryDirectory() as directory:
            fixtures = Path(directory)
            shutil.copyfile(FIXTURES / "reproduce_spec.py", fixtures / "reproduce_spec.py")
            vectors = json.loads((FIXTURES / "vectors.json").read_text(encoding="utf-8"))
            for message in vectors["messages"]:
                if message["name"] == "FeeDeclaration":
                    message["canonical_bytes_hex"] = "00"
                    break
            (fixtures / "vectors.json").write_text(json.dumps(vectors), encoding="utf-8")
            selected = reproduce("--message", "FeeDeclaration", fixtures=fixtures)
            other = reproduce("--message", "Heartbeat", fixtures=fixtures)

        self.assertEqual(selected.returncode, 1, selected.stderr)
        self.assertIn("[FAIL] FeeDeclaration", selected.stdout)
        self.assertIn("TOTAL CHECKS: 2  RESULT: MISMATCH", selected.stdout)
        self.assertEqual(other.returncode, 0, other.stderr)
        self.assertIn("TOTAL CHECKS: 2  RESULT: ALL MATCH", other.stdout)
