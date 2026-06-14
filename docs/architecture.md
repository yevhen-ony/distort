
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

 
## Object And Chunk Model

### Object

An Object is identified by `object_id`. It represents a collection of Chunks,
where each Chunk is labeled by a `chunk_key`. The Object does not contain payload bytes directly.
Payload bytes are stored as Chunks.

### Chunk Key

A `chunk_key` labels a Chunk within an Object.

The `chunk_key` is provided by the Client Layer. The Cluster does not interpret it or assign semantics to it.
Any meaning such as ordering, partitioning, naming, or task identity belongs to the Producer and Consumer.

The only Cluster-level requirement is that `chunk_key` values are unique within the same Object.

### Chunk

A Chunk is the stored byte unit. It is equipped with a generated `chunk_id`, size, and content digest.

Storage stores Chunk bytes by `chunk_id`.

### Metadata vs Bytes

Object metadata and Chunk bytes are separate but complementary.

- Object metadata describes how Chunks form an Object and is stored by the Master.
- Chunk bytes contain the payload and are stored on Storage instances.
- Reconstructing an Object requires both: metadata to know what to fetch, and Chunk bytes to rebuild the payload.

