from helpers import run_dos
from pathlib import Path
import pytest
import uuid


@pytest.fixture
def workdir(tmp_path: Path):
    return tmp_path


@pytest.fixture
def run_id():
    return uuid.uuid4().hex[:8]


@pytest.fixture
def created_objects():
    objects = []

    yield objects

    errors = []
    for object_id in objects:
        try:
            run_dos("object", "scale", object_id, "--replicas", "0")
        except AssertionError as err:
            errors.append(f"{object_id}: {err}")

    assert not errors, "cleanup failed:\n" + "\n".join(errors)
