package mutex

import (
	"hash/fnv"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
	"unsafe"

	constants "github.com/NikoMalik/low-level-functions/constants"
)

const spinCount = 100

type MutexExp struct {
	i int32
	_ [constants.CacheLinePadSize - unsafe.Sizeof(int32(0))]byte
}

func (m *MutexExp) get() int32 {
	return atomic.LoadInt32(&m.i)
}

func (m *MutexExp) set(i int32) {
	atomic.StoreInt32(&m.i, i)
}

func (m *MutexExp) Lock() {
	var spin int32
	for {
		if atomic.CompareAndSwapInt32(&m.i, 0, 1) {
			return
		}

		if spin < spinCount {
			spin++
			runtime.Gosched() // Yield the processor
		} else {
			time.Sleep(10 * time.Microsecond) // Back off a bit
			spin = 0
		}
	}
}

// Unlock releases the spinlock.
func (m *MutexExp) Unlock() {
	if m.get() == 0 {
		panic("BUG: Unlock of unlocked Mutex")
	}

	m.set(0)
}

type ShardedMutex struct {
	shards []MutexExp
	_      [constants.CacheLinePadSize - unsafe.Sizeof(MutexExp{})]byte
	count  int
	_      [constants.CacheLinePadSize - unsafe.Sizeof(int(0))]byte
}

func NewShardedMutex(shardCount int) *ShardedMutex {
	return &ShardedMutex{
		shards: make([]MutexExp, shardCount),
		count:  shardCount,
	}
}

func (s *ShardedMutex) GetShard(key int) *MutexExp {
	index := key % s.count
	return &s.shards[index]
}

func (s *ShardedMutex) Lock(key int) {
	shard := s.GetShard(key)
	shard.Lock()
}

func (s *ShardedMutex) Unlock(key int) {
	shard := s.GetShard(key)
	shard.Unlock()
}

func CalculateKey(data int) int {
	h := fnv.New32a()
	h.Write([]byte(strconv.Itoa(data)))
	return int(h.Sum32())
}
