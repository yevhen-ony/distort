# User Guide

This guide walks through a local DOS preview: starting a cluster, uploading and downloading an object,
inspecting placement, trying replication, observing metrics, and cleaning up. It assumes Docker Compose,
`make`, and basic command-line familiarity.


## Start The Cluster
---

The local preview environment uses Docker Compose. The `Makefile` wraps the common Docker commands,
so the guide uses `make` targets as the primary interface.

### Build The Service Images

```sh
make build
```

This builds:

- `dos-master:latest`: Master control-plane service.
- `dos-storage:latest`: Storage data-plane service.
- `dos-client:latest`: CLI environment used by the Client Layer.


### Start the Cluster

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


### Open A Client

In another terminal window, start the Client container:

```sh
make client
```

This starts an interactive shell in the Client container.
The container includes the `dos` CLI and uses the same Docker network as the Cluster.

The local `sandbox/` directory is mounted into the Client container at `/work/in`.
Use it to pass files between the host and the Client container.

### Sanity Check

In the Client container, verify that the Cluster is reachable:

```sh
dos system leader show
dos node list
```

The first command shows the active Master leader.
The second lists Storage instances currently known by the Master.
A healthy local preview should show one active Master and three registered Storage instances.


## Basic Upload/Download
---

The simplest way to use `dos` is to upload and download a file.


The Client splits a file into fixed-size Chunks and uploads them to the Cluster.
The Cluster represents the uploaded file as an Object addressed by `object_id`.

Downloading reverses the flow: the Client fetches the Object description,
downloads the required Chunks, and reconstructs the original file.


### Generate A File

You can use any local file and skip this step. For a reproducible first run,
create a sample file on the host because the Client container is intentionally minimal:

```sh
mkdir -p sandbox
dd if=/dev/urandom of=sandbox/sample.bin bs=1M count=100
```

This creates `sandbox/sample.bin` with size 100MB.

In the Client container, check that the file is visible:

```sh
ls ./in/
```

The default preview uses a `20MB` Chunk size, so a `100MB` file demonstrates a multi-Chunk upload.

### Upload An Object

In the Client container, run:

```sh
dos upload ./in/sample.bin --id sample
```

This uploads `./in/sample.bin` as an Object with object id `sample`.

The command prints uploaded Chunks, transferred bytes, and a final `Done` status.


### List Objects

In the Client container, list **Objects** currently known to the Cluster:

```sh
dos object list
```

The output should include the sample Object.
For each Object, it shows the number of attached Chunks and the replication target.


### Download An Object

In the Client container, download the Object back from the Cluster:

```sh
dos download sample --dest ./in/sample.out
```

The required argument is `object_id`; here it is `sample`.
The optional `--dest` flag writes the payload to a directory or file path.
If `--dest` points to a directory, `object_id` is used as the file name.


During download, the CLI shows transfer progress and finishes with status `Done`.

### Verify The Result

On the host, verify that the downloaded payload is identical to the original:

```sh
cmp sandbox/sample.bin sandbox/sample.out
```

`cmp` compares files byte by byte. If it exits without output, the files are identical.


## Inspect The Cluster
---
The Cluster stores an Object as distributed Chunks placed on Storage instances.
The commands below show how the uploaded `sample` Object is represented internally.

### Chunk List

List all **Chunks** currently known to the Cluster:

```sh
dos chunk list
```

The output shows each Chunk ID, size, actual replica count, and parent `object_id`.

Unlike `dos object list`, which shows the desired replication target,
`dos chunk list` shows actual replicas reported by Storage.
These values may differ temporarily while replication converges.


### List Storage

List active Storage instances registered in the Cluster:

```sh
dos node list
```

The CLI uses `node` to denote a Storage instance.

The output shows registered Storage instances, their addresses, the number of Chunks they currently store,
and used capacity.


### Describe Object

Describe the uploaded Object:

```sh
dos object describe sample
```

The output shows Object metadata and the Chunks attached to it.
This is the quickest way to understand how an Object is represented inside the Cluster.

### Describe Chunk

Similarly, you can describe a single Chunk by its `chunk_id`:

```sh
dos chunk describe <chunk_id>
```

Use one of the `chunk_id` values from `dos object describe sample` or `dos chunk list`.

The output shows Chunk metadata and the Storage instances that currently hold replicas.
This explains the replica count seen earlier from `dos chunk list`.

### Inspect Storage

Inspect a single Storage instance for a node-level view:

```sh
dos node inspect <storage-address>
```

The output shows used and total capacity, Cluster status, and held Chunk replicas.

## Replication
---

By default, the preview Cluster may use replication factor `1`.
Scale the sample Object to two replicas:

```sh
dos object scale sample --replicas 2
```

This updates the desired replication factor for the sample Object.
The target cannot exceed the number of active Storage instances.


Check the Object replication target:

```sh
dos object list
```

The Object list should show the updated target right away.
The Cluster reconciles Chunk replicas in the background, so actual replica counts may lag behind briefly.

Check the actual replica counts:

```sh
dos chunk list
```

With the default preview settings, replica counts usually converge in 3–5s.

To inspect placement after replication, describe the Object again:

```sh
dos object describe sample
```

### Simulate A Missing Storage Instance

Now simulate a Storage outage by pausing heartbeats for one Storage instance.

Use a `chunk_id` from the earlier `dos chunk list` output. If needed, list Chunks again,
pick one that belongs to the `sample` Object, and describe it:

```sh
dos chunk list
dos chunk describe <chunk_id>
```

The output shows Storage instances currently holding replicas of the Chunk.
Pick one Storage address from that output and pause its heartbeat:

```sh
dos node heartbeat <storage-address> --pause
```

This does not kill the container. It stops Storage liveness reports,
so the Master eventually treats the instance as unavailable.

Wait for the Cluster to react. With the default preview settings, this can take around 20s.

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

In the Client container, download the Object again:

```sh
dos download sample --dest ./in/sample.after-failure.out
```

On the host, verify the downloaded payload:

```sh
cmp sandbox/sample.bin sandbox/sample.after-failure.out
```

If `cmp` exits without output, the files are identical.


## Observability
---

The CLI is useful for direct checks: uploading, downloading, listing, describing,
and inspecting Cluster resources.

Some behavior is easier to understand over time. Replication, repair, request latency,
and throughput are all easier to evaluate with continuous signals instead of one command output.

For that, the project provides metrics and load tests.
Metrics show how the Cluster behaves while it is running. Load tests generate controlled activity,
so there is something meaningful to observe.

### Metrics

The Cluster exposes metrics for Master and Storage behavior.
They count low-level operations and measure durations, which helps evaluate Cluster health
and tune configuration for a particular use case.

The local preview starts Prometheus together with the Cluster. Open it in the browser:

```text
http://localhost:9090
```

This guide does not document individual metrics. The important point is that metrics are available
while you run the flows above, so you can observe uploads, replication changes, and Storage unavailability.

### Load Testing

Metrics are most useful when the Cluster is under controlled load.
The project includes a load-test tool available as `dos-load` inside the test container.

Each load-test job executes the same end-to-end workflow:

1. create an Object
2. upload a payload
3. scale Object replication
4. wait until replication converges
5. download the payload back
6. verify byte equality
7. clean up the Object

To try it out, build the test image that includes the additional Python-based tooling:

```sh
make build-test
```

Start the load-test shell:

```sh
make load
```

In the load-test container, run:

```sh
dos-load --workers 4 --duration 60 --size-mb 10
```

The main parameters are:

- `--workers`: number of concurrent workers.
- `--duration`: test duration in seconds.
- `--size-mb`: payload size per Object.

The load test prints per-job progress and a final report with succeeded/failed jobs, transferred data,
throughput, and average latency.

Here, latency means the duration of the whole workflow for one Object, not a single RPC or transfer operation.

Similarly, throughput means the amount of payload data that successfully completes the whole workflow per unit of time.
It is not the maximum raw write bandwidth the Cluster can accept.

Use the load test together with Prometheus to understand how the Cluster behaves under sustained activity.


## Cleanup
---

### Delete Object

The current CLI does not have a separate delete command. To remove an Object, set its replication target to `0`:

```sh
dos object scale sample --replicas 0
```

The Cluster reconciles this in the background by deleting Chunk replicas and then removing Object metadata.

You can check cleanup progress with:

```sh
dos object list
dos chunk list
```

### Stop The Cluster

Exit the Client or load-test shell if one is still open:

```sh
exit
```

Stop and remove the local Cluster:

```sh
make down
```
`make down` stops the Docker Compose services and removes containers, networks, and Compose-managed volumes for the dos project.

