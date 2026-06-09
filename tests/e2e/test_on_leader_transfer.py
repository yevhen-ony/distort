import uuid
from dataclasses import dataclass
from pathlib import Path
from tests.support.helpers import (
    assert_same_bytes,
    describe_object,
    write_random_file,
    download_object,
    upload_object,
    wait_object_replicated,
    scale_object,
    show_leader,
    change_leader,
    wait_nodes,
)


# @pytest.mark.skip()
def test_on_leader_transfer(workdir, run_id, cleanup):
    
    obj_before = gen_upload_wait(workdir, run_id, cleanup)

    # change_leader
    leader_before = show_leader()
    leader_after = change_leader() 
    assert leader_before["master_id"] != leader_after["master_id"]
    wait_object_replicated(obj_before.id)
    
    # old object accessible

    # describe object after leader change
    object_desc = describe_object(obj_before.id)
    assert object_desc["replication"] == 2
    
    for chunk in object_desc["chunks"]:
        assert len(chunk["sources"]) == 2

    # download object
    download_object(obj_before.id, str(obj_before.target))
    assert_same_bytes(obj_before.source, obj_before.target)

    # ensure nodes registered
    wait_nodes(3)

    # new object writable
    obj_after = gen_upload_wait(workdir, run_id, cleanup)
    
    # describe object after leader change
    object_desc = describe_object(obj_after.id)
    assert object_desc["replication"] == 2
    
    for chunk in object_desc["chunks"]:
        assert len(chunk["sources"]) == 2
    
    # download second object
    download_object(obj_after.id, str(obj_after.target))
    assert_same_bytes(obj_after.source, obj_after.target)


@dataclass
class UploadedObject:
    id: str
    source: Path
    target: Path

def gen_upload_wait(workdir, run_id, cleanup) -> UploadedObject:
    suffix = uuid.uuid4().hex[:8]
    object_id = f"e2e-on-leader-transfer-{run_id}-{suffix}"
    source =  workdir / f"src-{suffix}.bin"
    target =  workdir / f"res-{suffix}.bin"

    size = 10 * 1024 * 1024
    write_random_file(source, size)
    
    # upload & scale & wait
    upload_object(object_id, source)
    cleanup(lambda: scale_object(object_id, 0))
    scale_object(object_id, 2)
    wait_object_replicated(object_id)

    return UploadedObject(object_id, source, target)
