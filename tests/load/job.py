import time
from dataclasses import dataclass
from pathlib import Path
from tests.support.helpers import (
    assert_same_bytes,
    download_object,
    scale_object,
    upload_object,
    wait_object_replicated,
)

@dataclass
class JobResult:
    object_id: str
    size_bytes: int
    started_at: float
    finished_at: float
    elapsed_sec: float
    ok: bool
    error: str | None


    def __str__(self) -> str:
        status = "ok" if self.ok else "failed"
        size_mb = self.size_bytes / 1024 / 1024
        msg = (
            f"object_id={self.object_id} "
            f"size={size_mb}MB "
            f"status={status} "
            f"elapsed={self.elapsed_sec:.2f}s"
        )

        if self.error:
            msg += f" error={self.error}"

        return msg


def _run_job(object_id: str, source: Path, destination: Path):
    upload_object(object_id, source)
    scale_object(object_id, 2)
    wait_object_replicated(object_id)
    download_object(object_id, str(destination))
    assert_same_bytes(source, destination)


def run_job(object_id: str, source: Path, destination: Path) -> JobResult:
    started_at = time.monotonic()
    error = None
    ok = False

    try:
        _run_job(object_id, source, destination)
        ok = True

    except Exception as err:
        error = str(err) or repr(err)

    finally:
        try:
            scale_object(object_id, 0)
        except Exception as cleanup_err:
            if error is None:
                error = f"cleanup failed: {cleanup_err}"
            ok = False

        finished_at = time.monotonic()

    return JobResult(
        object_id=object_id,
        size_bytes=source.stat().st_size,
        started_at=started_at,
        finished_at=finished_at,
        elapsed_sec=finished_at - started_at,
        ok=ok,
        error=error,
    )
