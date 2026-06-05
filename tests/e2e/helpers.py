import json
import os
import subprocess
import time
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


def wait_until(check, timeout=15, interval=0.5, message="condition not met"):
    deadline = time.monotonic() + timeout
    last_error = None

    while time.monotonic() < deadline:
        try:
            if check():
                return
        except AssertionError as err:
            last_error = err

        time.sleep(interval)

    if last_error is not None:
        raise AssertionError(f"{message}: {last_error}")

    raise AssertionError(message)


def describe_chunk(chunk_id): 
    raw = run_dos_json("chunk", "describe", chunk_id)
    res = assert_success(raw, "describe_chunk")
    assert res["chunk_meta"]["chunk_id"] == chunk_id
    return res


def describe_object(object_id):
    raw = run_dos_json("object", "describe", object_id)
    res = assert_success(raw, "describe_object")
    assert res["object_id"] == object_id
    return res


def chunk_replicated(chunk_id, replica_count):
    res = describe_chunk(chunk_id)
    return len(res["sources"]) == replica_count 


def upload_object(object_id, source):
    raw = run_dos_json("upload", "--id", object_id, str(source))
    res = assert_success(raw, "object_transfer_progress")
    assert res["status"] == "Done"

    def cleanup():
        run_dos("object", "scale", object_id, "--replicas", "0")
    return cleanup


def download_object(object_id, destination):
    raw = run_dos_json("download", object_id, "--dest", destination)
    res = assert_success(raw, "object_transfer_progress")
    assert res["status"] == "Done"


def wait_object_replicated(object_id):

    desc_obj = describe_object(object_id)
    replica_count = desc_obj["replication"]

    for chunk in desc_obj["chunks"]:
        chunk_id = chunk["chunk_meta"]["chunk_id"]
        wait_until(
            lambda: chunk_replicated(chunk_id, replica_count),
            message=f"chunk {chunk_id} was not replicated on time",
        )

def scale_object(object_id, replica_count):
    run_dos("object", "scale", object_id, "--replicas", str(replica_count))

