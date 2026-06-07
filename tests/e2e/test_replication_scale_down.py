from tests.support.helpers import (
    assert_same_bytes,
    describe_object,
    write_random_file,
    download_object,
    upload_object,
    wait_object_replicated,
    scale_object,
)

def test_replication_scale_down(workdir, run_id, cleanup):
    object_id = f"e2e-replication-scale-down-{run_id}"

    source = workdir / "source.bin"
    downloaded = workdir / "downloaded.bin"

    size = 10 * 1024 * 1024
    write_random_file(source, size)

    # upload object
    upload_object(object_id, source)
    cleanup(lambda: scale_object(object_id, 0))
    scale_object(object_id, 2)
    wait_object_replicated(object_id)

    # ensure replication 
    object_desc = describe_object(object_id)
    assert object_desc["replication"] == 2
    
    for chunk in object_desc["chunks"]:
        assert len(chunk["sources"]) == 2

    # scale down
    scale_object(object_id, 1)
    wait_object_replicated(object_id)
    
    # ensure replication 
    object_desc = describe_object(object_id)
    assert object_desc["replication"] == 1
    
    for chunk in object_desc["chunks"]:
        assert len(chunk["sources"]) == 1

    # download object
    download_object(object_id, str(downloaded))
    assert_same_bytes(source, downloaded)
