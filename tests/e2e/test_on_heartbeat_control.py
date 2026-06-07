from tests.support.helpers import (
    inspect_node,
    list_nodes,
    pause_node,
    resume_node,
    wait_node_paused,
    wait_node_resumed,
)


def test_on_heartbeat_control(cleanup):

    # list nodes before pause 
    nodes_before = list_nodes()
    count_before = len(nodes_before) 
    assert count_before > 0

    node = nodes_before[0]

    # pause node
    pause_node(node["address"])
    cleanup(lambda: resume_node(node["address"]))
    
    # inspect paused node
    inspect_paused = inspect_node(node["address"])
    assert inspect_paused["heartbeat"]["status"] == "paused"

    # wait node paused
    wait_node_paused(node["address"])
    
    # list nodes after pause 
    nodes_after = list_nodes() 
    count_after = len(nodes_after) 
    assert count_after < count_before

    # paused node not listed
    assert node["address"] not in [item["address"] for item in nodes_after]

    # resume node
    resume_node(node["address"])
    
    # inspect resumed node
    inspect_resumed = inspect_node(node["address"])
    assert inspect_resumed["heartbeat"]["status"] == "running"

    # wait node resumed
    wait_node_resumed(node["address"])

    # list nodes after pause 
    nodes_final = list_nodes()
    count_final = len(nodes_final) 
    assert count_final == count_before

    # resumed node listed
    assert node["address"] in [item["address"] for item in nodes_final]
