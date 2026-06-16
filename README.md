# Distributed Object Store

> Working name: `dos`
>
> Status: **preview**. This repo is intended for local demos, experimentation, and hands-on exploration.

This project is an experimental distributed object delivery and storage system.

It takes an abstract Object, splits it into Chunks, stores those Chunks across remote
Storage instances, and later reconstructs the Object by fetching its Chunks back.


## Documentation
---

- [User Guide](docs/user-guide.md): run the local preview, upload/download Objects, try replication, and inspect the Cluster.
- [Architecture](docs/architecture.md): understand the system model, roles, data flow, and design choices.


## Terminology
---

- **Cluster**: the system boundary: Master control plane plus Storage data plane.
- **Master**: the control plane role; coordinates metadata, placement, Storage state, and replication intent.
- **Storage**: the data plane role; keeps Chunk bytes locally and serves uploads/downloads.
- **Client**: external tooling or code that creates, uploads, downloads, and reconstructs Objects.
- **Producer**: anything that creates Object data or individual Chunks.
- **Consumer**: anything that fetches Object data or individual Chunks.
- **Compute**: a future worker role that may produce or consume Chunks near Storage.
- **Object**: a logical unit of data addressed by `object_id`.
- **Chunk**: a piece of an Object, identified within the Object by a chunk key.

```text
Cluster = Master + Storage
```


## What Is This?
---

This is a Cluster for moving and storing chunked Objects.

An Object can be a file, a dataset, compute output, or a task split into pieces.
The Cluster does not need to know what the Object means. It only knows how to place,
track, move, replicate, fetch, and reconstruct its Chunks.


## Why It Exists
---

Distributed systems often produce data in one place and consume it somewhere else.

A Client may upload a file today. Later, Compute may produce Chunks directly near Storage.
A Consumer should still be able to fetch the whole Object by `object_id`, without knowing
where each Chunk was created or stored.

The project explores this idea: storage is not just a passive bucket, but a delivery
fabric between Producers, Storage, Compute, and Consumers.

A normal object-storage flow:   `Client -> Cluster -> Client`

A decoupled compute flow:       `Producer / Compute -> Cluster -> Consumer`

A task-distribution flow:       `Client splits task -> Cluster -> Compute`


## How It Works
---

The system is composed of two internal **Cluster** roles and one external **Client**:

- **Master** coordinates metadata, placement, Storage state, and replication intent.
Multiple Master instances form a Raft-backed control plane, with one active Master
coordinating the Cluster.

- **Storage** persists Chunk bytes locally and serves direct uploads/downloads.
Adding Storage instances expands capacity for new Chunks without affecting existing data.

- **Client** runs outside the Cluster and handles Object creation, upload, download, and reconstruction.
  A Client can be both Producer and Consumer.

### Upload / Produce

1. A Producer creates an Object with the Master.
2. The Producer splits the Object into Chunks.
3. The Master allocates Chunk slots assigned to the Object.
4. The Producer uploads Chunk bytes to Storage.
5. Storage persists Chunk bytes locally and reports inventory back to the Master.
6. The Master tracks which Storage instances actually hold each Chunk.

### Download / Consume

1. A Consumer asks the Master where the Object’s Chunks are available.
2. The Consumer downloads the Chunks directly from Storage.
3. The Consumer reconstructs the Object from the downloaded Chunks.

Splitting is intentionally delegated to the Producer, because chunking can be
domain-specific. **Client** tooling handles the common mechanics: creating Objects,
allocating Chunks, uploading Chunks to Storage, downloading Chunks, and reconstructing
Objects.

Replication is handled as a background concern. The Master tracks desired replication
intent, Storage reports actual Chunk inventory, and the Cluster can schedule repairs
when Chunks are under-replicated.


## Why Chunks?
---

Chunks make large or abstract Objects operational.

A Chunk is small enough to move, retry, verify, place, replicate, and report independently.
This gives the Cluster a bounded unit of work instead of forcing every operation to
reason about the whole Object at once.

Chunking also makes streaming natural: Producers upload data piece by piece,
Storage persists and serves pieces independently, and Consumers download Chunks and
reconstruct the Object without requiring one monolithic transfer.


## Implementation
---

`dos` is a Go-based distributed object store using gRPC, Docker Compose,
Helm manifests, and end-to-end tests.

It supports object upload/download, manual chunk workflows, Storage reporting,
and chunk replication.

