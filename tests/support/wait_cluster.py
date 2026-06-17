import argparse
from tests.support.helpers import (
    list_nodes,
    show_leader,
    wait_until,
)


def wait_for_leader(timeout, interval):
    def check():
        leader = show_leader()
        assert leader.get("master_id"), "leader response has no master_id"
        return True

    print("waiting for leader election...", flush=True)
    wait_until(
        check,
        timeout=timeout,
        interval=interval,
        message="leader was not elected",
    )
    print("leader elected!", flush=True)


def wait_for_nodes(expected_count, timeout, interval):
    def check():
        nodes = list_nodes()
        actual_count = len(nodes)
        assert actual_count >= expected_count, (
            f"registered nodes: {actual_count}/{expected_count}"
        )
        return True

    print(f"waiting for {expected_count} registered nodes...", flush=True)
    wait_until(
        check,
        timeout=timeout,
        interval=interval,
        message="not enough registered nodes",
    )
    print("storage nodes registered!", flush=True)


def main():
    parser = argparse.ArgumentParser(description="Wait until the cluster is ready")
    parser.add_argument("--nodes", type=int, default=3)
    parser.add_argument("--leader-timeout", type=float, default=60)
    parser.add_argument("--nodes-timeout", type=float, default=60)
    parser.add_argument("--interval", type=float, default=1)
    args = parser.parse_args()

    wait_for_leader(args.leader_timeout, args.interval)
    wait_for_nodes(args.nodes, args.nodes_timeout, args.interval)
    print("cluster is ready!", flush=True)


if __name__ == "__main__":
    main()

