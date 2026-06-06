from helpers import (
    allocate_chunk,
    assert_same_bytes,
    create_object,
    pull_chunk,
    push_chunk,
    scale_object,
    trigger_report,
    write_random_file,
)


def test_chunk_flow(workdir, run_id, cleanup):
    object_id = f"e2e-manual-chunk-flow-{run_id}"
    chunk_key = "part-0001"

    source = workdir / "source.bin"
    downloaded = workdir / "downloaded.bin"

    write_random_file(source, 1024 * 1024)

    # create empty object
    create_object(object_id)
    cleanup(lambda: scale_object(object_id, 0))

    # allocate chunk
    alloc_chunk = allocate_chunk(object_id, chunk_key)

    chunk_id = alloc_chunk["chunk_id"]
    target = alloc_chunk["targets"][0]

    # push chunk bin
    push_chunk(chunk_id, target["node_id"], target["address"], source)

    # force node to report
    trigger_report(target["address"], chunk_id)

    # pull chunk bin
    pull_chunk(chunk_id, target["node_id"], target["address"], str(downloaded))

    assert_same_bytes(source, downloaded)
