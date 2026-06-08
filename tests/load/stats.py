from dataclasses import dataclass
from worker import WorkerResult

@dataclass
class LoadResult:
    worker_count: int
    elapsed_sec: float
    success_count: int
    error_count: int
    success_bytes: int
    total_bytes: int
    avg_op_sec: float
    success_bytes_per_sec: float
    workers: list[WorkerResult]


    def __str__(self) -> str:
        return "\n".join([
            "REPORT:",
            "==============================",
            f"workers:          {self.worker_count}",
            f"elapsed:          {self.elapsed_sec:.2f} s",
            "------------------------------",
            f"jobs failed:      {self.error_count}",
            f"jobs succeeded:   {self.success_count}",
            "------------------------------",
            f"transferred:      {mb(self.success_bytes):.2f} MB",
            f"throughput:       {mb(self.success_bytes_per_sec):.2f} MB/s",
            f"latency:          {self.avg_op_sec:.2f} s",
            "==============================",
        ])


def mb(value: int | float) -> float:
    return value / 1024 / 1024

def summarize_load(workers: list[WorkerResult]) -> LoadResult:
    elapsed_sec = max((worker.elapsed_sec for worker in workers), default=0.0)
    job_elapsed_sec = sum(worker.elapsed_sec for worker in workers)

    success_count = sum(worker.success_count for worker in workers)
    error_count = sum(worker.error_count for worker in workers)

    success_bytes = sum(worker.success_bytes for worker in workers)
    total_bytes = sum(worker.total_bytes for worker in workers)

    total_count = success_count + error_count

    if total_count > 0 and elapsed_sec > 0:
        avg_op_sec = job_elapsed_sec / total_count 
        success_bytes_per_sec = success_bytes / elapsed_sec
    else:
        avg_op_sec = 0.0
        success_bytes_per_sec = 0.0

    return LoadResult(
        worker_count=len(workers),
        elapsed_sec=elapsed_sec,
        success_count=success_count,
        error_count=error_count,
        success_bytes=success_bytes,
        total_bytes=total_bytes,
        avg_op_sec=avg_op_sec,
        success_bytes_per_sec=success_bytes_per_sec,
        workers=workers,
    )
