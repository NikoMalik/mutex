package mutex

import (
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	constants "github.com/NikoMalik/low-level-functions/constants"
)

const spinCount = 200

const (
	mutexUnlocked = 0
	mutexLocked   = 1

	mutexWrite          = 1 << 0
	mutexReadOffset     = 1 << 1
	mutexUnderflow      = ^int32(mutexWrite)
	mutexWriterUnset    = ^int32(mutexWrite - 1)
	mutexReaderDecrease = ^int32(mutexReadOffset - 1)
)

var spin int16

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
	if m == nil {
		panic("BUG: Lock of nil Mutex")
	}
loop:
	for {
		if atomic.CompareAndSwapInt32(&m.i, mutexUnlocked, mutexLocked) {
			return
		}

		if spin < spinCount {
			spin++
			runtime.Gosched() // Yield the processor
			goto loop
		} else {
			time.Sleep(1 * time.Microsecond) // Back off a bit
			spin = mutexUnlocked
		}
	}
}

func (m *MutexExp) Unlock() {
	if m == nil {
		panic("BUG: Unlock of nil Mutex")
	}

	state := m.get()
	if state == mutexUnlocked {
		panic("BUG: Unlock of unlocked Mutex")
	}

	if state&mutexWrite == mutexWrite {
		m.set(mutexUnlocked)
	} else {
		// The lock is held in read mode; decrease the reader count
		atomic.AddInt32(&m.i, -mutexReadOffset)
	}
}

func (m *MutexExp) RLock() {
loop:
	for {
		state := atomic.LoadInt32(&m.i)

		if state&mutexWrite == 0 {
			// Lock is free or held in read mode; increment reader count
			if atomic.CompareAndSwapInt32(&m.i, state, state+mutexReadOffset) {
				return
			}
		}

		if spin < spinCount {
			spin++
			runtime.Gosched() // Yield the processor
			goto loop
		} else {
			time.Sleep(1 * time.Microsecond) // Back off a bit
			spin = mutexUnlocked
		}
	}
}

func (m *MutexExp) RUnlock() {
	state := atomic.LoadInt32(&m.i)
	if state&mutexWrite != 0 {
		panic("BUG: RUnlock of locked Mutex")
	}

	if state == mutexUnlocked {
		panic("BUG: RUnlock of unlocked RWMutex")
	}

	if atomic.AddInt32(&m.i, -mutexReadOffset) == mutexUnlocked {
		return
	}
}

type ShardedMutex struct {
	shards []MutexExp
	_      [constants.CacheLinePadSize - unsafe.Sizeof(&MutexExp{})]byte
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
	if s == nil {
		panic("BUG: GetShard of nil ShardedMutex")
	}
	return &s.shards[key&(len(s.shards)-1)]
}
func (s *ShardedMutex) Lock(key int) {
	if s == nil {
		panic("BUG: Lock of nil ShardedMutex")
	}
	s.GetShard(key).Lock()
}

func (s *ShardedMutex) Unlock(key int) {
	if s == nil {
		panic("BUG: Unlock of nil ShardedMutex")
	}
	s.GetShard(key).Unlock()
}
func CalculateKey(data int) int {
	x := uint32(data)
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 3
	x ^= x >> 17
	x ^= x << 5
	return int(x)
}
