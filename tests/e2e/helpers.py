import json
import os
import subprocess
from pathlib import Path


def run_dos(*args):
    cmd = ["dos", *args]
    proc = subprocess.run(
        cmd,
        cwd="/work",
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    if proc.returncode != 0:
        raise AssertionError(
            f"command failed: {' '.join(cmd)}\n"
            f"stdout:\n{proc.stdout}\n"
            f"stderr:\n{proc.stderr}"
        )

    return proc.stdout


def run_dos_json(*args):
    out = run_dos(*args, "-o", "json")
    return json.loads(out)


def write_random_file(path: Path, size: int):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(os.urandom(size))


def assert_same_bytes(left: Path, right: Path):
    assert left.read_bytes() == right.read_bytes()


def assert_success(envelope, operation: str):
    assert envelope["operation"] == operation
    assert "error" not in envelope, envelope.get("error")
    assert "result" in envelope
    return envelope["result"]
