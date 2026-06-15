# Architecture

This document describes the architecture of the project at the conceptual level.

The goal is to explain the main entities, responsibilities, flows, and assumptions needed to understand
the idea behind the system. The terms used here are architectural concepts, not necessarily one-to-one
mappings to code objects or single running processes.

Some concepts may be implemented by several cooperating components, and some implementation details may
combine multiple concepts. The purpose of this document is to provide an overall model of the project,
not an exhaustive implementation reference.


## System Model

An **Object** is a collection of Chunks plus metadata that describes how those Chunks form the Object.

The system provides Object construction, storage, and delivery by storing Chunk bytes in a distributed
way while keeping Object metadata separately. Chunk bytes carry the data; Object metadata makes it
possible to locate the Chunks and reconstruct the Object.

The system is organized into two architectural parts: the **Cluster** and the **Client Layer**.

The **Cluster** is the shared runtime of the system. It maintains the Object catalog, stores Chunk bytes,
and tracks who those Chunks can be accessed.

The **Client Layer** provides protocol-compliant access to the Cluster. It runs with Producers and Consumers,
and turns application data into Object and Chunk operations.

### Cluster Roles

**Master** is the control-plane role. It maintains Cluster metadata to coordinate Chunk allocation, placement,
lookup, and replication targets. A Cluster may run multiple Master instances; they form a Raft group with one
active leader under healthy quorum, while the other Masters act as followers.

**Storage** is the data-plane role. It stores Chunk bytes locally and reports its inventory to the Master.
Storage instances are elastic: starting another Storage instance is enough to add capacity to the Cluster.

### Communication Paths

One useful way to describe the system is to separate three kinds of traffic:

- **Operational Flow**: metadata, coordination, and commands.
- **Data Flow**: Chunk bytes.
- **Consensus Flow**: raft specific Master-to-Master coordination.

Operational Flow always goes through the Master. The Master is where Cluster decisions are made:
which Objects exist, which Chunk slots belong to an Object, where Chunks should be placed,
where existing Chunks can be found, and which Storage instances should participate in repair.

```text
Client Layer -> Master   # create Object, allocate Chunk, describe Object
Storage      -> Master   # register Storage, report stored Chunks
Master       -> Storage  # request replication or deletion
```

Data Flow carries Chunk bytes. The Master does not participate in byte transfer directly;
data transfer takes place directly between the Client Layer and Storage, or between Storage instances.

```text
Client Layer -> Storage       # upload Chunk bytes
Storage      -> Client Layer  # download Chunk bytes
Storage      -> Storage       # copy Chunk bytes for replication or repair
```

Consensus Flow keeps Master instances coordinated.

```text
Master <-> Master  # raft coordination and leadership
```

**The key design property** is that metadata and decisions flow through the Master, while Chunk bytes move
directly between the parties that produce, store, copy, or consume them.

 
## System Entities 

### Object

An Object is identified by `object_id`. It represents a collection of Chunks,
where each Chunk is labeled by a `chunk_key`. The Object does not contain payload bytes directly.
Payload bytes are stored as Chunks.

The `chunk_key` is provided by the Client Layer. The Cluster does not interpret it or assign semantics to it.
Any meaning such as ordering, partitioning, naming, or task identity belongs to the Producer and Consumer.
The only Cluster-level requirement is that `chunk_key` values are unique within the same Object.

### Chunk

A Chunk is the stored byte unit. It is equipped with a generated `chunk_id`, size, and content digest.

Storage instances store Chunk bytes by `chunk_id` and report inventory information about stored Chunks to the Master.


### Metadata vs Bytes

Object metadata and Chunk bytes are separate but complementary.

- Object metadata describes how Chunks form an Object and is stored by the Master.
- Chunk bytes contain the payload and are stored on Storage instances.
- Reconstructing an Object requires both: metadata to know what to fetch, and Chunk bytes to rebuild the payload.

### Chunk Placement

Chunk Placement contains the set of Storage instances holding a Chunk.

The system uses Chunk Placement to locate Chunk bytes, reason about replication state, and coordinate access, repair, or deletion.


## Core Flows

The main flows are Produce and Consume. In both cases, the Master coordinates Object and Chunk metadata, while Chunk bytes move directly through Storage.

### Produce

1. A Producer defines an Object and splits its payload into Chunks.
2. The Client Layer creates the Object through the Master.
3. For each Chunk, the Client Layer provides a `chunk_key` and asks the Master to allocate a Chunk.
4. The Master records the Chunk as part of the Object and returns Storage targets.
5. The Client Layer uploads Chunk bytes directly to Storage.
6. Storage stores the Chunk and reports it to the Master.
7. The Master updates its view of where the Chunk is available.

### Consume

1. A Consumer asks the Master for Object metadata.
2. The Master returns the Object’s Chunks and their available Storage locations.
3. The Client Layer downloads Chunk bytes directly from Storage.
4. The Client Layer reconstructs the Object from the downloaded Chunks.

### Replication / Repair

Clients can define a replication factor for a given Object. Replication is handled as a background flow.

1. The Master tracks the replication target for the Object.
2. Storage instances report which Chunks they store.
3. The Master compares Chunk Placement with the replication target and sends a reconciliation request to Storage when they differ.
4. Storage either deletes an extra replica or copies Chunk bytes to another Storage instance.
5. Storage instances report updated Chunk Placement back to the Master.

Replication changes Chunk Placement; it does not change Object metadata or Chunk content.

Replication uses a chain request. The Master sends one reconciliation request to a Storage instance with an ordered list
of target Storage instances. That Storage instance copies the Chunk to the next target and forwards the remaining request.
The process continues until the target list is exhausted.

```text
Master -> Storage A: replicate Chunk X to [B, C]
Storage A -> Storage B: copy Chunk X, forward [C]
Storage B -> Storage C: copy Chunk X
```

Chain replication decentralizes the repair work. The Master selects the target chain, but Storage instances execute and
forward the replication request. This distributes transfer load across Storage instances instead of concentrating it on
one source or on the Master.

The **tradeoff** is reduced central control: the Master observes the result through reports rather than driving every
copy step directly. If a chain fails, Storage reports the failure and the Master can schedule reconciliation again.


## State Model

The system separates authoritative, observed, and derived states.

- **Authoritative state** is the declared state initiated by the Client Layer and accepted by the Cluster.
It expresses the requirements the Cluster tries to satisfy. It must be persisted and synchronized between
the Cluster participants responsible for maintaining it.

- **Observed state** is state reported by Cluster participants about their current local condition.
It describes what currently exists or is available, not what the Cluster requires. The receiver can cache it,
but should be able to refresh or rebuild it from later reports.

- **Derived state** is an optimization built from authoritative and observed state. It does not add independent
facts and can be dropped and rebuilt without affecting consistency.

### Object Catalog

The Master maintains a view of all Objects created through the Client Layer and stores that view as the Object Catalog.
The Object Catalog is authoritative state.

The Object Catalog is required to reconstruct payloads from independently stored Chunks. It defines Chunk membership,
Chunk labels, and the replication target for each Object.

This information is composed by the Client Layer, recorded by the Master, and acts as the primary source of truth for the Cluster.

In the multi-Master profile, Object Catalog changes are synchronized through Raft before they are accepted,
so leadership can move to another Master without losing accepted catalog state.

### Storage Inventory

Each Storage instance maintains local knowledge of the Chunks it stores and reports that information to the Master.

Storage Inventory is observed state: it describes the current physical presence of Chunk replicas on Storage instances.
It does not define Object structure or replication targets; those come from the Object Catalog.

Committed Chunk bytes are the payload itself. Storage Inventory only reports where those bytes currently exist.

On startup, a Storage instance scans its local disk, rebuilds its local inventory, and reports discovered Chunks to the Master.
During normal operation, the inventory is kept as an in-memory cache and updated as Chunks are uploaded, deleted, or replicated.

For a Storage instance, the local disk is the source of truth across restart or failure events.

### Chunk Placement

Chunk Placement is derived state.

It is built from Object Catalog entries and Storage Inventory reports. For each Chunk known to the Object Catalog,
the Master maintains a view of which Storage instances currently report holding it.

When serving an Object description, the Master returns the full Chunk Placement for each Chunk. This lets the Client Layer
choose which Storage instance to read from and gives it alternative sources if one Storage instance is unavailable.

Accumulated Chunk Placement also gives the Master a storage map: a view of where data is stored and which Storage instances
have available capacity. The Master can use this view to choose targets for new Chunks, avoid overloaded Storage instances,
and coordinate replication or deletion.

This creates a path toward rebalancing: as Storage capacity changes, the Master can use placement information to move replicas
without changing Object metadata or Chunk content.

### State Summary

| State | Kind | Maintained by | Recovery expectation |
| --- | --- | --- | --- |
| Object Catalog | Authoritative | Master | Persisted and synchronized by the control plane |
| Storage Inventory | Observed | Storage, reported to Master | Rebuilt from local Storage disk and reports |
| Chunk Placement | Derived | Master | Rebuilt from Object Catalog and Storage Inventory |

## Failure Model

Durability and replication claims are meaningful only together with a failure model. This section describes the assumptions
used by the current architecture.

### Assumptions

* The system assumes non-malicious participants. Master identity and Storage identity are trusted by configuration today;
stronger authentication such as mTLS is future work.

* Clients may be incorrect or interrupted. They may allocate Chunks without uploading bytes, retry operations,
upload invalid data, or attempt to upload bytes for unknown Chunk IDs. The Cluster validates what it can and treats
incomplete or inconsistent data as non-committed.

* Storage instances are volatile from the Cluster point of view: they may join, leave, restart, timeout, or become
temporarily unreachable. Storage reports are observations and may become stale.

* Chunk bytes may be corrupted in transit or at rest. Chunks are identified with digests, and inconsistent bytes
are rejected when validated.

* Multi-Master mode assumes a healthy Raft quorum and a single legitimate leader. Without quorum, the control plane
cannot safely accept authoritative state changes.

### Covered Behavior

- A Chunk with at least one reachable valid replica can be downloaded and verified by digest.
- If one Storage replica is unavailable during download, the Client Layer can try another Storage instance from Chunk Placement.
- After replication converges, a Chunk with replication target `R` can survive the loss of up to `R - 1` Storage replicas,
assuming at least one valid replica remains reachable.
- If Storage restarts with its local disk intact, it can rebuild inventory from disk and report stored Chunks back to the Master.
- If a Client allocates a Chunk but does not upload valid bytes, the Chunk remains part of the Object with zero valid replicas.
- If Storage reports bytes that conflict with the expected digest, the report is rejected or the bytes are rejected by
the Client Layer during download.
- If the active Master fails while Raft quorum remains healthy, another Master can become leader and continue from accepted catalog state.

### Outside This Model

- Malicious participants intentionally violating the protocol.
- Loss of all valid replicas of a Chunk.
- Loss of Master quorum.
- Restart durability for deployments that do not persist the required Master or Storage state.
- Strong guarantees about repair time under arbitrary capacity pressure or network instability.


## Design Tradeoffs

### Client-Owned Chunk Semantics

The Cluster does not interpret `chunk_key` values or Chunk content. This keeps the Object model generic:
the same Cluster can store files, datasets, compute outputs, or task partitions.

The tradeoff is that Producers and Consumers must agree on Chunk semantics. The Cluster can enforce uniqueness of
`chunk_key` values within an Object, but it cannot know whether the chosen keys are meaningful for a particular application.

### Direct Data Path

Chunk bytes move directly between the Client Layer and Storage, or between Storage instances during replication.
The Master coordinates metadata and placement, but it does not proxy payload bytes.

This reduces load on the Master and keeps the control plane focused on coordination. The tradeoff is that the Client Layer
must handle transfer retries, source selection, and reconstruction behavior.

### Report-Driven Placement

The Master learns Chunk Placement from Storage reports. This keeps Storage autonomous and allows instances to restart,
rebuild local inventory, and report what they actually have.

The tradeoff is that Placement can be temporarily stale. The system must tolerate stale observations and reconcile them
through later reports and repair.

### Chain Replication

Replication is executed as a chain of Storage-to-Storage transfers. The Master chooses the target chain,
while Storage instances copy bytes and forward the request.

This decentralizes replication load and keeps the Master out of the data path. The tradeoff is less direct control over
each copy step; failures are reported back and reconciled by later repair attempts.


## Current Implementation

The current implementation is written in Go and uses gRPC for service communication.

It provides three binaries: Master, Storage, and Client CLI. Docker Compose and Helm manifests are available for local
and Kubernetes-style deployments, and end-to-end tests cover Object flow, Chunk flow, replication, and node-loss scenarios.

The current preview profile favors fast iteration over production persistence defaults. Production deployment profiles should
explicitly configure durable Master state and durable Storage volumes according to the intended failure model.

### Observability And Load Testing

The system exposes metrics for Master and Storage behavior, including transfer, placement, replication, and health signals.

Load tests complement these metrics by exercising the Cluster under controlled pressure and producing useful data about
throughput, latency, and replication behavior.


## Future Direction

The architecture is designed to support Compute as another Producer or Consumer role.

A Compute worker could produce Chunks directly near Storage, allowing Consumers to fetch the resulting Object by `object_id`.
In the reverse direction, a Client could split a task into Chunks, store them as an Object, and let Compute workers consume
those Chunks as work items.

Future work includes locality-aware placement, production persistence profiles, rebalancing, richer placement policies,
and a stable Client Layer API.

Locality-aware placement would allow Storage targets to be selected using parameters such as rack, subnet, region,
or proximity to Compute workers. This would make it possible to keep data close to where it is produced or consumed.

