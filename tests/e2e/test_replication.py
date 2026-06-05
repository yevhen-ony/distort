from helpers import (
    assert_same_bytes,
    describe_object,
    write_random_file,
    download_object,
    upload_object,
    wait_object_replicated,
    describe_chunk,
    scale_object,
)


def test_replication(workdir, run_id, cleanup):
    object_id = f"e2e-replication-{run_id}"

    source = workdir / "source.bin"
    downloaded = workdir / "downloaded.bin"

    size = 10 * 1024 * 1024
    write_random_file(source, size)

    # upload object
    delete_obj = upload_object(object_id, source)
    cleanup(delete_obj)

    wait_object_replicated(object_id)

    # scale object replication
    scale_object(object_id, 2)

    # describe object after scale
    object_desc_after = describe_object(object_id)
    assert object_desc_after["replication"] == 2
    
    wait_object_replicated(object_id)

    for chunk in object_desc_after["chunks"]:
        chunk_id = chunk["chunk_meta"]["chunk_id"]
        desc_chunk = describe_chunk(chunk_id)
        assert len(desc_chunk["sources"]) == 2

    # download object
    download_object(object_id, str(downloaded))

    assert_same_bytes(source, downloaded)
