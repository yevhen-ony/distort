from helpers import (
    run_dos_json,
    write_random_file,
    assert_same_bytes,
    assert_success,
)


def test_manual_chunk_flow(workdir, run_id):
    object_id = f"e2e-object-{run_id}"
    chunk_key = "part-0001"

    source = workdir / "source.bin"
    downloaded = workdir / "downloaded.bin"

    write_random_file(source, 1024 * 1024)

    create_raw = run_dos_json("object", "create", object_id)
    assert_success(create_raw, "create_object")

    alloc_raw = run_dos_json("chunk", "allocate", object_id, "--key", chunk_key)
    alloc = assert_success(alloc_raw, "allocate_chunk")

    chunk_id = alloc["chunk_id"]
    target = alloc["targets"][0]

    push_raw = run_dos_json(
      "chunk", "push", str(source),
      "--id", chunk_id,
      "--node-id", target["node_id"],
      "--node-addr", target["address"],
    )
    assert_success(push_raw, "push_chunk")

    report_raw = run_dos_json(
      "node", "report", target["address"],
      "--chunk", chunk_id,
    )
    report = assert_success(report_raw, "trigger_report")
    assert chunk_id in report["scheduled"]

    got_raw = run_dos_json(
      "chunk", "get", chunk_id,
      "--node-id", target["node_id"],
      "--node-addr", target["address"],
      "--dest", str(downloaded),
    )
    assert_success(got_raw, "download_chunk")
    assert_same_bytes(source, downloaded)
