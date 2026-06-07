from tests.support.helpers import (
    write_random_file,
    upload_object,
    scale_object,
    wait_object_replicated,
    describe_object,
    pause_node,
    wait_node_paused,
    list_nodes,
    resume_node,
    download_object,
    assert_same_bytes,
)

def test_reconstruction_after_node_loss(workdir, run_id, cleanup):
    object_id = f"e2e-reconstruction-{run_id}"

    source = workdir / "source.bin"
    downloaded = workdir / "downloaded.bin"

    size = 10 * 1024 * 1024
    write_random_file(source, size)
    
    # upload & scale & wait
    upload_object(object_id, source)
    cleanup(lambda: scale_object(object_id, 0))
    scale_object(object_id, 3)
    wait_object_replicated(object_id)

    # ensure chunks replicated
    desc = describe_object(object_id)
    for chunk in desc["chunks"]:
        assert len(chunk["sources"]) == 3
    
    # select sources of the first chunk
    sources = desc["chunks"][0]["sources"]
    node1 = sources[0]
    node2 = sources[1]

    # fisrt node lost
    pause_node(node1["address"])
    cleanup(lambda: resume_node(node1["address"]))
    wait_node_paused(node1["address"])

    # second node lost
    pause_node(node2["address"])
    cleanup(lambda: resume_node(node2["address"]))
    wait_node_paused(node2["address"])
    
    # ensure single node left
    node_addrs = [item["address"] for item in list_nodes()]
    assert node1["address"] not in node_addrs
    assert node2["address"] not in node_addrs
    assert len(node_addrs) == 1
    
    download_object(object_id, str(downloaded))
    assert_same_bytes(source, downloaded)
    
