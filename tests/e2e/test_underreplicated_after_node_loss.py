from helpers import (
    describe_object,
    list_nodes,
    pause_node,
    resume_node,
    scale_object,
    upload_object,
    wait_node_paused,
    wait_object_replicated,
    write_random_file,
)


def test_underreplicated_after_node_loss(workdir, run_id, cleanup):
    object_id = f"e2e-underreplicated-after-node-loss-{run_id}"

    source = workdir / "source.bin"

    size = 10 * 1024 * 1024
    write_random_file(source, size)

    # upload & scale & wait 
    upload_object(object_id, source)
    cleanup(lambda: scale_object(object_id, 0))
    scale_object(object_id, 3)
    wait_object_replicated(object_id)

    # ensure replication 
    object_desc = describe_object(object_id)
    assert object_desc["replication"] == 3
    
    for chunk in object_desc["chunks"]:
        assert len(chunk["sources"]) == 3
    
    # pause node & wait
    node = list_nodes()[0]
    pause_node(node["address"])
    cleanup(lambda: resume_node(node["address"]))
    wait_node_paused(node["address"])
    
    # ensure desired replication 
    object_desc = describe_object(object_id)
    assert object_desc["replication"] == 3
    
    # observer real replication
    for chunk in object_desc["chunks"]:
        # stays underreplicated
        assert len(chunk["sources"]) == 2
        
        # paused node is not listed 
        node_ids = [src["node_id"] for src in chunk["sources"]]
        assert node["node_id"] not in node_ids

