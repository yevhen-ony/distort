from tests.support.helpers import (
    assert_same_bytes,
    describe_object,
    download_object,
    list_nodes,
    pause_node,
    resume_node,
    scale_object,
    upload_object,
    wait_node_paused,
    wait_object_replicated,
    write_random_file,
)


def test_replication_heal(workdir, run_id, cleanup):
    object_id = f"e2e-replication-heal-{run_id}"

    source = workdir / "source.bin"
    downloaded = workdir / "downloaded.bin"

    size = 10 * 1024 * 1024
    write_random_file(source, size)

    # upload  & scale & wait 
    upload_object(object_id, source)
    cleanup(lambda: scale_object(object_id, 0))
    scale_object(object_id, 2)
    wait_object_replicated(object_id)

    # describe object
    obj = describe_object(object_id)
    assert obj["replication"] == 2
    assert len(obj["chunks"]) > 0
    
    # ensure all chunks replicated
    for chunk in obj["chunks"]:
        assert len(chunk["sources"]) == 2

    # pick node to pause
    selected_node = obj["chunks"][0]["sources"][0]

    # pause node
    pause_node(selected_node["address"]) 
    cleanup(lambda: resume_node(selected_node["address"]))
    wait_node_paused(selected_node["address"])

    # ensure node is not listed
    assert selected_node["node_id"] not in [
        item["node_id"] for item in list_nodes()
    ]

    # wait object replicated
    wait_object_replicated(object_id)

    # describe object
    obj = describe_object(object_id)
    assert obj["replication"] == 2
    assert len(obj["chunks"]) > 0
    
    # ensure all chunks replicated
    for chunk in obj["chunks"]:
        assert len(chunk["sources"]) == 2

        node_ids = [item["node_id"] for item in chunk["sources"]]
        assert selected_node["node_id"] not in node_ids

    download_object(object_id, str(downloaded))
    assert_same_bytes(source, downloaded)
