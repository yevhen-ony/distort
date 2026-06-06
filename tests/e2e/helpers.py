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


def pause_node(node_addr):
    pause_raw = run_dos_json("node", "heartbeat", node_addr, "--pause")
    pause_res = assert_success(pause_raw, "heartbeat_control")
    assert pause_res["address"] == node_addr 
    assert pause_res["heartbeat"]["status"] == "paused"


def resume_node(node_addr):
    resume_raw = run_dos_json("node", "heartbeat", node_addr, "--resume")
    resume_res = assert_success(resume_raw, "heartbeat_control")
    assert resume_res["address"] == node_addr 
    assert resume_res["heartbeat"]["status"] == "running"


def is_node_listed(node_addr):
    list_node_raw = run_dos_json("node", "list")
    list_node_res = list_node_raw["result"]

    node_addrs = [node_item["address"] for node_item in list_node_res]
    return node_addr in node_addrs 


def wait_node_paused(node_addr):
    # wait master reaction
    wait_until(
        lambda: not is_node_listed(node_addr),
        message=f"node {node_addr} is still listed",
    )

    inspect_paused = inspect_node(node_addr)
    assert inspect_paused["heartbeat"]["status"] == "paused"


def wait_node_resumed(node_addr):
    # wait master reaction
    wait_until(
        lambda: is_node_listed(node_addr),
        message=f"node {node_addr} is still not listed",
    )

    inspect_resumed = inspect_node(node_addr)
    assert inspect_resumed["heartbeat"]["status"] == "running"


def list_nodes():
    list_raw = run_dos_json("node", "list")
    list_res = assert_success(list_raw, "list_nodes")
    return list_res


def inspect_node(node_addr):
    inspect_raw = run_dos_json("node", "inspect", node_addr)
    inspect_res = assert_success(inspect_raw, "inspect_node")
    return inspect_res


def list_objects():
    list_raw = run_dos_json("object", "list")
    list_res = assert_success(list_raw, "list_objects")
    return list_res


def list_chunks():
    list_raw = run_dos_json("chunk", "list")
    list_res = assert_success(list_raw, "list_chunks")
    return list_res


def create_object(object_id):
    create_raw = run_dos_json("object", "create", object_id)
    assert_success(create_raw, "create_object")


def allocate_chunk(object_id, chunk_key):
    alloc_raw = run_dos_json("chunk", "allocate", object_id, "--key", chunk_key)
    alloc_res = assert_success(alloc_raw, "allocate_chunk")
    return alloc_res


def push_chunk(chunk_id, node_id, node_addr, source):
    push_raw = run_dos_json(
        "chunk", "push", str(source),
        "--id", chunk_id,
        "--node-id", node_id,
        "--node-addr", node_addr, 
    )
    assert_success(push_raw, "push_chunk")

def pull_chunk(chunk_id, node_id, node_addr, dest):
    pull_chunk_raw = run_dos_json(
        "chunk", "get", chunk_id,
        "--node-id", node_id,
        "--node-addr", node_addr,
        "--dest", dest,
    )
    assert_success(pull_chunk_raw, "download_chunk")

def trigger_report(node_addr, chunk_id):
    trigger_report_raw = run_dos_json(
        "node", "report", node_addr,
        "--chunk", chunk_id,
    )
    trigger_report_res = assert_success(trigger_report_raw, "trigger_report")
    assert chunk_id in trigger_report_res["scheduled"]

