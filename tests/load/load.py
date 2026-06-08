#!/usr/bin/env python3

import argparse
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor
from worker import run_worker
from stats import summarize_load 


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--workers", type=int, default=1)
    parser.add_argument("--duration", type=int, default=60)
    parser.add_argument("--size-mb", type=int, default=10)
    parser.add_argument("--workdir", default="/tmp/dos-load")
    args = parser.parse_args()

    if args.workers <= 0:
        raise ValueError("--workers must be positive")
    if args.duration <= 0:
        raise ValueError("--duration must be positive")
    if args.size_mb <= 0:
        raise ValueError("--size-mb must be positive")

    workdir = Path(args.workdir)
    workdir.mkdir(parents=True, exist_ok=True)

    size_bytes = args.size_mb * 1024 * 1024

    with ThreadPoolExecutor(max_workers=args.workers) as pool:
        futures = [
            pool.submit(run_worker, worker_id, args.duration, size_bytes, workdir)
            for worker_id in range(args.workers)
        ]

        results = []
        for future in futures:
            results.append(future.result())


    load_result = summarize_load(results)
    print(load_result, flush=True)


if __name__ == "__main__":
    main()
