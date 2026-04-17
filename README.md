# Minimal Distributed Object Store (Learning Project)

## Goal

Build a simplified distributed object store inspired by GFS.

Primary requirement:

> A client uploads a large object. The object remains downloadable as long as at least one replica of each chunk is available.

---

## Architecture

Components:

- **Master**
  - stores metadata only
  - tracks objects, chunks, and replica locations
  - tracks alive chunkservers (heartbeat)

- **ChunkServer**
  - stores raw chunk data on disk
  - serves upload/download requests

- **Client**
  - splits objects into chunks
  - uploads/downloads data
  - interacts with master for metadata

---

## Data Model

### File

- `object_id`
- `object_size`
- `chunks: map[index] -> chunk_id`

### Chunk

- `chunk_id`
- `index` (position in object)
- `replicas: []chunkserver`

---

## Key Design Decisions

- **Client-side chunking**
  - client splits object into chunks (fixed size, e.g. 4MB)
  - last chunk may be smaller

- **Chunks are raw bytes**
  - no semantics
  - no object awareness in chunkservers

- **Explicit chunk index**
  - client assigns `index`
  - order defined by index, not upload order

- **Replication**
  - each chunk stored on multiple chunkservers (e.g. 3)

- **Master decides placement**
  - assigns which chunkservers store each chunk

---

## Communication

Use **gRPC** for all communication.

### Master service (planned)

- `StartFile`
- `AllocateChunk`
- `GetFile`
- `Heartbeat`

### Chunk service

- `PutChunk`
- `GetChunk`

---

## Upload Flow (streaming)

1. Client starts object upload
2. For each chunk:
   - split locally
   - request placement from master
   - upload to chunkservers
3. (Completion semantics TBD: commit or implicit)

---

## Download Flow

1. Client requests object layout from master
2. Master returns:
   - ordered chunk indices
   - replica locations
3. Client downloads chunks in order
4. If replica fails → try another

---

## Invariants

- chunk indices are **unique per object**
- chunkservers store **only (chunk_id → bytes)**
- master stores **no object data**
- object reconstruction = **sort by index + concatenate**

---

## Open Questions (intentionally postponed)

- explicit `CommitFile` vs implicit completion
- failure recovery / re-replication
- checksums / validation
- partial reads
- rebalancing

