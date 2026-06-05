from helpers import (
    assert_same_bytes,
    assert_success,
    run_dos_json,
    write_random_file,
    describe_object,
    upload_object,
    download_object,
    wait_object_replicated,
)


def test_object_flow(workdir, run_id, cleanup):
    object_id = f"e2e-object-flow-{run_id}"

    source = workdir / "source.bin"
    downloaded = workdir / "downloaded.bin"

    size = 10 * 1024 * 1024  # 10 MB
    write_random_file(source, size)
    
    delete_obj_fn = upload_object(object_id, source)
    cleanup(delete_obj_fn)

    # list objects
    list_objects_raw = run_dos_json("object", "list")
    list_objects_res = assert_success(list_objects_raw, "list_objects")
    object_ids = [object_item["object_id"] for object_item in list_objects_res]
    assert object_id in object_ids

    # describe object
    object_desc = describe_object(object_id) 
    assert len(object_desc["chunks"]) > 0

    # list chunks
    list_chunks_raw = run_dos_json("chunk", "list")
    list_chunks_res = assert_success(list_chunks_raw, "list_chunks")
    listed_chunks = {chunk["chunk_id"]: chunk for chunk in list_chunks_res}

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
