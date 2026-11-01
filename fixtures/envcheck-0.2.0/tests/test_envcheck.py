import os
import subprocess
import sys
import tempfile
import unittest

from envcheck.rc import RcError, parse

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def run(args, env=None, cwd=None):
    full = dict(os.environ)
    full["PYTHONPATH"] = ROOT
    for key in ("ALPHA", "BETA", "GAMMA"):
        full.pop(key, None)
    if env:
        full.update(env)
    return subprocess.run(
        [sys.executable, "-m", "envcheck.cli"] + args,
        capture_output=True,
        text=True,
        env=full,
        cwd=cwd or ROOT,
    )


class TestParse(unittest.TestCase):
    def test_keeps_file_order(self):
        self.assertEqual(parse("BETA\nALPHA\n"), ["BETA", "ALPHA"])

    def test_skips_blank_and_comment(self):
        self.assertEqual(parse("# note\n\nALPHA\n"), ["ALPHA"])

    def test_deduplicates_keeping_first(self):
        self.assertEqual(parse("ALPHA\nBETA\nALPHA\n"), ["ALPHA", "BETA"])

    def test_rejects_space(self):
        with self.assertRaises(RcError):
            parse("ALPHA BETA\n")


class TestCli(unittest.TestCase):
    def rc(self, body):
        handle = tempfile.NamedTemporaryFile("w", suffix=".rc", delete=False)
        handle.write(body)
        handle.close()
        self.addCleanup(os.unlink, handle.name)
        return handle.name

    def test_all_present(self):
        path = self.rc("ALPHA\n")
        result = run(["--rc", path], env={"ALPHA": "1"})
        self.assertEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "all 1 required variables present\n")

    def test_reports_each_missing(self):
        path = self.rc("ALPHA\nBETA\n")
        result = run(["--rc", path])
        self.assertEqual(result.returncode, 1)
        self.assertEqual(
            result.stdout,
            "missing: ALPHA\nmissing: BETA\n2 of 2 required variables missing\n",
        )

    def test_unreadable_rc_exits_two(self):
        result = run(["--rc", "/nonexistent/path/.envcheckrc"])
        self.assertEqual(result.returncode, 2)
        self.assertIn("cannot read", result.stderr)

    def test_bad_rc_exits_two(self):
        path = self.rc("ALPHA BETA\n")
        result = run(["--rc", path])
        self.assertEqual(result.returncode, 2)
        self.assertIn("contains a space", result.stderr)

    def test_unexpected_argument_exits_two(self):
        result = run(["extra"])
        self.assertEqual(result.returncode, 2)
        self.assertIn("unexpected argument", result.stderr)


if __name__ == "__main__":
    unittest.main()
