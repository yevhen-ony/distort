
An **Object** is a collection of Chunks plus metadata that describes how those Chunks form the Object.

The system provides Object construction, storage, and delivery by storing Chunk bytes in a distributed
way while keeping Object metadata separately. Chunk bytes carry the data; Object metadata makes it
possible to locate the Chunks and reconstruct the Object.

### System Model

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

