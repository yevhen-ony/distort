# User Guide


## Start The Cluster 
---

The local preview environment uses Docker Compose. The `Makefile` wraps the common Docker commands,
so the guide uses `make` targets as the primary interface.

### Step 1. Build the service images:

```sh
make build
```

This builds:

- `dos-master:latest`: Master control-plane service.
- `dos-storage:latest`: Storage data-plane service.
- `dos-client:latest`: CLI environment used by the Client Layer.


### Step 2. Start the Cluster:

```sh
make up
```

It runs the root Docker Compose file and deploys the Cluster into a Docker network.
By default, the preview starts three Master instances forming a Raft-backed control plane,
three Storage instances, and Prometheus for metrics.

This topology is also the baseline expected by the end-to-end tests.

In the logs, expect to see Raft leader election first.
During startup, Storage instances may temporarily log errors about failing to reach the active Master.
This is expected before Raft elects a leader and the discovery path becomes available.

After a leader is elected, Storage instances register and start sending regular heartbeat
messages to the active Master. These messages may appear as recurring debug logs.


### Step 3. Open A Client

In another terminal window:

```sh
make client
```

This starts an interactive shell in the Client container.
The container includes the dos CLI and uses the same Docker network as the Cluster.

The local `sandbox/` directory is mounted into the Client container at `/work/in`.
Use it to pass files between the host and the Client container.

### Sanity Check

Inside the Client container, verify that the Cluster is reachable:

```sh
dos system leader show
dos node list
```

The first command shows the active Master leader.
The second lists Storage instances currently known by the Master.
A healthy local preview should show one active Master and three registered Storage instances.


## Basic Upload/Downlad 
---

The simplest way to use `dos` is to upload and download a file.

By default, the Client takes a given file, splits it into fixed-size parts called
Chunks, and uploads those Chunks to the Cluster. The uploaded file is represented
in the Cluster as an Object and addressed by `object_id`.

Downloading performs the reverse operation: the Client fetches the Object description,
downloads the required Chunks, and reconstructs the original file.


### Step 0. Generate A File

You can use any local file for this flow and skip this step. For the first try,
follow the instructions below so the guide stays reproducible.

The Client container is minimal and contains the `dos` CLI, but not the usual shell toolbox.
To generate the sample file, open a third terminal window on the host machine and run:

```sh
mkdir -p sandbox
dd if=/dev/urandom of=sandbox/sample.bin bs=1M count=100
```

This creates a file called sample.bin with size 100MB in the sandbox/ directory.

Go back to the Client terminal and check that the file is visible:

```sh
ls ./in/
```

With the default preview configuration, the Client uses a `20MB` Chunk size,
so a `100MB` file is large enough to demonstrate a general multi-Chunk flow.


### Step 1. Upload An Object 

From the Client terminal, run:

```sh
dos upload ./in/sample.bin --id sample
```

This uploads the file at `./in/sample.bin` to the Cluster as an Object and assigns it the object id `sample`.

The command prints upload progress.
It lists the Chunks being uploaded and the number of bytes transferred to the Cluster.
When the upload completes successfully, the final status should be `Done`.


### Step 3. List Objects

List **Objects** currently known to the Cluster:

```sh
dos object list
```

The output should include the sample Object.
For each Object, it shows the number of attached Chunks and the replication target.

### Step 3. Download An Object 

Download the Object back from the Cluster:

```sh
dos download sample --dest ./in/sample.out
```

The required argument is `object_id`; here it is `sample`.

The optional `--dest` flag specifies where to write the downloaded payload.
It can point either to an existing directory or to a file path that does not exist yet.
If `--dest` points to a directory, the object_id will be used as the file name.

During download, the CLI shows transfer progress. A successful download finishes with status Done.


### Step 4. Verify The Result

To ensure that the downloaded payload is identical to the original, run from the host terminal:

```sh
cmp sandbox/sample.bin sandbox/sample.out
```

`cmp` compares files byte by byte. If it exits without output, the files are identical.


## Inspect The Cluster
---

The Cluster stores an Object as multiple distributed Chunks placed on Storage instances.
The CLI exposes this internal state through several administrative commands.

We already used `dos object list` to see Objects known to the Cluster.
In this section, we look at the other inspection commands and connect them to the uploaded `sample` Object.

### Chunk List

List all **Chunks** currently known to the Cluster:

```sh
dos chunk list
```

The output shows Chunks created for uploaded Objects, including their Chunk IDs,
size, actual replica count, and parent object_id.

In contrast to the Object list, where replication is the desired state set by the user,
the Chunk list reveals the actual replica count reported by Storage.
These values may differ temporarily while replication converges.


### List Storage

It is often useful to list all active Storage instances registered in the Cluster:

```sh
dos node list
```

Note that the CLI uses node to denote a Storage instance.

The output shows registered Storage instances, their addresses, the number of Chunks they currently store,
and used capacity.


### Describe Object

During the lifetime of an Object, it is often useful to summarize its current state:

```sh
dos object describe sample
```

The output provides metadata about the Object itself, as well as the list of Chunks attached to the Object.

This is the most useful command for understanding how an Object is currently represented inside the Cluster.

### Describe Chunk

Similarly, you can describe a single Chunk by its `chunk_id`:

```sh
dos chunk describe <chunk_id>
```

Use one of the chunk_id values from `dos object describe sample` or `dos chunk list`.

The output provides Chunk metadata and shows where this Chunk is currently available in the Cluster.
From here, you can see which Storage instances hold replicas of the Chunk and naturally derive the
replica count exposed as a single number by `dos chunk list`.

### Inspect Storage

For a full diagnostic of the Cluster, inspecting a single Storage instance is indispensable.

Use an address from `dos node list`:

```sh
dos node inspect <storage-address>
```

The output shows the used and total capacity of the Storage instance,
its status within the Cluster, and the list of Chunk replicas it holds.


## Replication

By default, the preview Cluster may use replication factor `1`.
You can change the replication target for an Object with:

```sh
dos object scale sample --replicas 2
```

The current implementation does not allow the replication target to exceed the number of
active Storage instances. This limitation may be lifted in the future.

This updates the desired replication factor for the sample Object.

Check the Object replication target:
```sh
dos object list
```

The Object list should show the updated replication target right away.

The Cluster reconciles Chunk replicas in the background, so actual replica counts may not catch up to the target immediately.

Check the actual replica counts:

```sh
dos chunk list
```

After some time, it should show the updated replica counts.
The exact convergence time depends on configuration, but with the default preview settings it is usually around 3-5s.

To inspect placement after replication, describe the Object again:

```sh
dos object describe sample
```

### Simulate A Missing Storage Instance

For the sake of experiment, freeze one Storage instance that holds one of the Chunks.

Use a chunk_id from the earlier `dos chunk list` output. If you did not save one, list Chunks again,
pick one that belongs to the `sample` object and describe it:

```sh
dos chunk list
dos chunk describe <chunk_id>
```

The output shows Storage instances currently holding replicas of the Chunk.
Pick one Storage address from that output and pause its heartbeat:

```sh
dos node heartbeat <storage-address> --pause
```

This does not kill the container. It makes the Storage instance stop reporting liveness,
so the Master eventually treats it as unavailable.

Wait for the Cluster to react. With the default preview configuration, this can take around 20s.

List active Storage instances again:
```sh
dos node list
```

One Storage instance should be missing from the list. This imitates a failed or unreachable Storage instance.

List Chunks again:

```sh
dos chunk list
```

Depending on timing, you may briefly see affected Chunks with a lower actual replica count.
After repair converges, the actual replica count should return to the replication target.

The Object should remain downloadable:

```sh
dos download sample --dest ./in/sample.after-failure.out
```

Verify the downloaded payload from the host terminal:
```sh
cmp sandbox/sample.bin sandbox/sample.after-failure.out
```

If cmp exits without output, the files are identical.
