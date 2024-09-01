# MutexExp and ShardedMutex Documentation

## Overview

This package provides two mutex implementations: `MutexExp` and `ShardedMutex`. These implementations are designed for efficient locking in concurrent applications where minimizing contention is critical.

### MutexExp

`MutexExp` is a simple spinlock implementation that avoids the overhead associated with traditional blocking mutexes. It is designed for situations where the lock is held for a very short time and where spinning is more efficient than context switching.

### ShardedMutex

`ShardedMutex` is a composite mutex that distributes locks across multiple "shards" or smaller mutexes (`MutexExp`). This design reduces contention by allowing multiple operations to proceed in parallel as long as they target different shards. It's particularly useful in scenarios where you have a large number of independent keys or resources that need to be locked concurrently.

---

## MutexExp

### Fields

- `i int32`: The internal state of the mutex. A value of `0` indicates the mutex is unlocked, while `1` indicates it is locked.
- `_[CacheLinePadSize]byte`: Padding to avoid false sharing between processors. This ensures that the `MutexExp` structure occupies its own cache line, reducing contention on multi-core systems.

### Methods

- `get() int32`: Safely retrieves the current state of the mutex using atomic operations.
- `set(i int32)`: Safely sets the state of the mutex using atomic operations.
- `Lock()`: Attempts to acquire the lock. If the lock is not immediately available, it spins for a short period before yielding the processor and, eventually, sleeping to reduce CPU usage.
- `Unlock()`: Releases the lock. If the lock is already unlocked, it panics, indicating a logic error in the program.

### Advantages

- **Low overhead**: Spinlocks can be more efficient than traditional mutexes in low-contention scenarios where locks are held for a very short duration.
- **No context switching**: Since `MutexExp` avoids blocking, it can be more efficient in high-performance applications where context switching would add significant overhead.

### Risks

- **CPU usage**: In high-contention scenarios, `MutexExp` can cause excessive CPU usage as threads spin while waiting for the lock.
- **Potential deadlock**: If `Unlock` is not called correctly, it can result in a deadlock, especially since there is no way to detect misuse at compile time.
- **Limited use cases**: Spinlocks are best suited for short critical sections. If the critical section takes too long, it can negate the benefits of using a spinlock.

---

## ShardedMutex

### Fields

- `shards []MutexExp`: An array of `MutexExp` instances. Each shard can be locked independently, allowing multiple threads to operate on different shards concurrently.
- `count int`: The number of shards. Determines the level of granularity for locking.

### Methods

- `NewShardedMutex(shardCount int) *ShardedMutex`: Initializes a new `ShardedMutex` with the specified number of shards.
- `GetShard(key int) *MutexExp`: Computes the shard index based on the given key and returns the corresponding `MutexExp` instance.
- `Lock(key int)`: Acquires the lock for the shard corresponding to the given key.
- `Unlock(key int)`: Releases the lock for the shard corresponding to the given key.

### Advantages

- **Reduced contention**: By sharding the mutex, `ShardedMutex` reduces the chance of contention, allowing multiple threads to lock different shards concurrently.
- **Scalability**: This approach scales well with the number of threads, as the probability of contention decreases with the number of shards.

### Risks

- **Hash collisions**: If the keys map to the same shard, the benefits of sharding are reduced. Proper key distribution is crucial to avoid hot spots.
- **Complexity**: The use of sharded locks introduces additional complexity in managing and ensuring that the correct shard is locked for a given operation.
- **Partial locking**: If multiple resources are accessed that map to different shards, there is a risk of deadlock if locks are not acquired and released in a consistent order.

---

## Usage Scenarios

### MutexExp

- **Low-latency applications**: Where locking time is very short, and the overhead of context switching must be avoided.
- **High-performance code**: Where avoiding the overhead of blocking mutexes can result in significant performance gains.

### ShardedMutex

- **Concurrent data structures**: Ideal for data structures like hash tables, where different keys can be locked independently.
- **High-contention environments**: Where traditional locking mechanisms would result in excessive contention and reduced performance.

---
