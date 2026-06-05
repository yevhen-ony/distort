from helpers import (
    run_dos,
    run_dos_json,
    write_random_file,
    assert_same_bytes,
    assert_success,
)


def test_chunk_flow(workdir, run_id, cleanup):
    object_id = f"e2e-manual-chunk-flow-{run_id}"
    chunk_key = "part-0001"

    source = workdir / "source.bin"
    downloaded = workdir / "downloaded.bin"

    write_random_file(source, 1024 * 1024)

    # create empty object
    create_object_raw = run_dos_json("object", "create", object_id)
    assert_success(create_object_raw, "create_object")

    cleanup(lambda: run_dos("object", "scale", object_id, "--replicas", "0"))

    # allocate chunk
    alloc_chunk_raw = run_dos_json("chunk", "allocate", object_id, "--key", chunk_key)
    alloc_chunk_res = assert_success(alloc_chunk_raw, "allocate_chunk")

    chunk_id = alloc_chunk_res["chunk_id"]
    target = alloc_chunk_res["targets"][0]

    # push chunk bin
    push_raw = run_dos_json(
      "chunk", "push", str(source),
      "--id", chunk_id,
      "--node-id", target["node_id"],
      "--node-addr", target["address"],
    )
    assert_success(push_raw, "push_chunk")

    # force node to report
    trigger_report_raw = run_dos_json(
      "node", "report", target["address"],
      "--chunk", chunk_id,
    )
    trigger_report_res = assert_success(trigger_report_raw, "trigger_report")
    assert chunk_id in trigger_report_res["scheduled"]

    download_chunk_raw = run_dos_json(
      "chunk", "get", chunk_id,
      "--node-id", target["node_id"],
      "--node-addr", target["address"],
      "--dest", str(downloaded),
    )
    assert_success(download_chunk_raw, "download_chunk")
    assert_same_bytes(source, downloaded)
