from helpers import (
    run_dos_json,
    wait_until,
    assert_success,
)

def test_on_heartbeat_control(cleanup):

    # list nodes before pause 
    before_list_raw = run_dos_json("node", "list")
    before_list_res = assert_success(before_list_raw, "list_nodes")
    count_before = len(before_list_res) 
    assert count_before > 0

    node = before_list_res[0]

    # pause node
    pause_raw = run_dos_json("node", "heartbeat", node["address"], "--pause")
    pause_res = assert_success(pause_raw, "heartbeat_control")
    assert pause_res["address"] == node["address"]
    assert pause_res["heartbeat"]["status"] == "paused"

    # resume node on cleanup
    cleanup(lambda: run_dos_json("node", "heartbeat", node["address"], "--resume"))

    # inspect paused node
    inspect_paused_raw = run_dos_json("node", "inspect", node["address"])
    inspect_paused_res = assert_success(inspect_paused_raw, "inspect_node")
    assert inspect_paused_res["heartbeat"]["status"] == "paused"

    def is_node_count_changed(previous_count):
        list_raw = run_dos_json("node", "list")
        list_res = list_raw["result"]
        count = len(list_res)
        return count != previous_count

    # wait master reaction
    wait_until(
        lambda: is_node_count_changed(count_before),
        message="node count did not changed after heartbeat control",
    )

    # list nodes after pause 
    after_list_raw = run_dos_json("node", "list")
    after_list_res = assert_success(after_list_raw, "list_nodes")
    count_after = len(after_list_res) 
    assert count_after < count_before

    # paused node not listed
    node_ids = [node_item["node_id"] for node_item in after_list_res]
    assert node["node_id"] not in node_ids

    # resume node
    resume_raw = run_dos_json("node", "heartbeat", node["address"], "--resume")
    resume_res = assert_success(resume_raw, "heartbeat_control")
    assert resume_res["address"] == node["address"]
    assert resume_res["heartbeat"]["status"] == "running"

    # inspect resumed node
    inspect_resumed_raw = run_dos_json("node", "inspect", node["address"])
    inspect_resumed_res = assert_success(inspect_resumed_raw, "inspect_node")
    assert inspect_resumed_res["heartbeat"]["status"] == "running"
    
    # wait master reaction
    wait_until(
        lambda: is_node_count_changed(count_after),
        message="node count did not changed after heartbeat control",
    )

    # list nodes after pause 
    final_list_raw = run_dos_json("node", "list")
    final_list_res = assert_success(final_list_raw, "list_nodes")
    count_final = len(final_list_res) 
    assert count_final == count_before
