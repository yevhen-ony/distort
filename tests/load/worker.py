import time
import uuid
import itertools
from pathlib import Path
from job import JobResult, run_job
from dataclasses import dataclass

from tests.support.helpers import (
    write_random_file,
)
 
@dataclass
class WorkerResult:
    worker_id: int

    elapsed_sec: float
    success_count: int
    error_count: int
    success_bytes: int
    total_bytes: int

    success_ops_per_sec: float 
    total_ops_per_sec: float
    success_bytes_per_sec: float
    total_bytes_per_sec: float

    jobs: list[JobResult]


def summarize_worker(
    worker_id: int,
    jobs: list[JobResult],
) -> WorkerResult:

    elapsed_sec = sum(job.elapsed_sec for job in jobs)

    success_count = sum(1 for job in jobs if job.ok)
    error_count = len(jobs) - success_count
    total_count = len(jobs)

    success_bytes = sum(job.size_bytes for job in jobs if job.ok)
    total_bytes = sum(job.size_bytes for job in jobs)

    if elapsed_sec > 0:
        success_ops_per_sec = success_count / elapsed_sec
        total_ops_per_sec = total_count / elapsed_sec
        success_bytes_per_sec = success_bytes / elapsed_sec
        total_bytes_per_sec = total_bytes / elapsed_sec
    else:
        success_ops_per_sec = 0.0
        total_ops_per_sec = 0.0
        success_bytes_per_sec = 0.0
        total_bytes_per_sec = 0.0

    return WorkerResult(
        worker_id=worker_id,

        elapsed_sec=elapsed_sec,

        success_count=success_count,
        error_count=error_count,

        success_bytes=success_bytes,
        total_bytes=total_bytes,

        success_ops_per_sec=success_ops_per_sec,
        total_ops_per_sec=total_ops_per_sec,
        success_bytes_per_sec=success_bytes_per_sec,
        total_bytes_per_sec=total_bytes_per_sec,

        jobs=jobs,
    )


def run_worker(
    worker_id: int,
    duration_sec: int,
    size_bytes: int,
    workdir: Path,
) -> WorkerResult:
    deadline = time.monotonic() + duration_sec 
    counter = itertools.count()
    results = []

    suffix = uuid.uuid4().hex[:6]
    source = workdir / f"worker-w{worker_id}-{suffix}.src"
    write_random_file(source, size_bytes)

    while time.monotonic() < deadline:
        i = next(counter)
        object_id = f"load-w{worker_id}-{suffix}-i{i}"
        destination = workdir / f"{object_id}.dst"

        result = run_job(object_id, source, destination)
    
        destination.unlink(missing_ok=True)

        print(result, flush=True)
        results.append(result)

    return summarize_worker(worker_id, results)
