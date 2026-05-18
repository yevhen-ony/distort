##  ISSUE: Replication Thrashing Under Stale State and Capacity Pressure

###  What 

Master retries replication faster than storage can report updated replica state, while per-node heavy-op quota is tight. This creates repeated replica chain failed / re-schedule loops for the same chunk, even though convergence eventually happens.

###  Why

  - MaxParallelHeavyOps too low causes temporary service busy / hop failures.
  - ReplicationInterval too short compared to storage report visibility delay means master acts on stale actual replica count.

###  To Do 

  1. Set MaxParallelHeavyOps >= max replication factor (prefer +1).
  2. Set ReplicationInterval close to storage report delay (target around P95 report latency).
  3. Keep chain-failure reporting enabled, but add scheduler cooldown/backoff per chunk to avoid retry thrash.
  4. Keep queue dedup/coalescing so repeated failures do not spawn duplicate work.

###  Result 

  Lower control-plane noise, fewer redundant retries, and stable eventual convergence to desired replica count.
