from tests.support.helpers import (
    assert_same_bytes,
    write_random_file,
    describe_object,
    upload_object,
    download_object,
    wait_object_replicated,
    list_objects,
    scale_object,
    list_chunks,
)


def test_object_flow(workdir, run_id, cleanup):
    object_id = f"e2e-object-flow-{run_id}"

    source = workdir / "source.bin"
    downloaded = workdir / "downloaded.bin"

    size = 10 * 1024 * 1024  # 10 MB
    write_random_file(source, size)
    
    upload_object(object_id, source)
    cleanup(lambda: scale_object(object_id, 0))

    # ensure object listed 
    objects = list_objects()
    assert object_id in [item["object_id"] for item in objects]

    # ensure object with chunks 
    object_desc = describe_object(object_id) 
    assert len(object_desc["chunks"]) > 0

    # list chunks
    chunks = list_chunks()
    listed_chunks = {item["chunk_id"]: item for item in chunks}

    # ensure chunks listed
    for chunk in object_desc["chunks"]:
        # chunk is listed
        chunk_id = chunk["chunk_meta"]["chunk_id"]
        assert chunk_id in listed_chunks

        # listed chunk belongs to object
        chunk_info = listed_chunks[chunk_id]
        assert chunk_info["object_id"] == object_id

    wait_object_replicated(object_id)
    
    # download object
    download_object(object_id, str(downloaded))

    # compare bytes
    assert_same_bytes(source, downloaded)
