package mutex

import (
	"sync"
	"testing"
	"time"
)

func BenchmarkMutexExp(b *testing.B) {
	var m MutexExp
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m.Lock()

				time.Sleep(1 * time.Millisecond)
				m.Unlock()
			}
		}()
	}
	wg.Wait()
}

func BenchmarkStandardMutex(b *testing.B) {
	var m sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m.Lock()
				// Critical section
				time.Sleep(1 * time.Millisecond)
				m.Unlock()
			}
		}()
	}
	wg.Wait()
}

func BenchmarkShardedMutex(b *testing.B) {
	var m *ShardedMutex = NewShardedMutex(64)
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := CalculateKey(j)
				m.Lock(key)
				// Critical section
				time.Sleep(1 * time.Millisecond)
				m.Unlock(key)
			}
		}()
	}
	wg.Wait()
}

func BenchmarkMutex_Concurrent(b *testing.B) {
	const goroutines = 100
	var mu sync.Mutex
	counter := 0
	wg := sync.WaitGroup{}

	wg.Add(goroutines)
	b.ResetTimer()

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()

			current := counter
			time.Sleep(10 * time.Millisecond)
			counter = current + 1
		}()
	}

	wg.Wait()

	if counter != goroutines {
		b.Fatalf("expected counter to be %d, got %d", goroutines, counter)
	}
}

func BenchmarkMutexExp_Concurrent(b *testing.B) {
	const goroutines = 100
	var mu MutexExp
	counter := 0
	wg := sync.WaitGroup{}

	wg.Add(goroutines)

	b.ResetTimer()

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()

			current := counter
			time.Sleep(10 * time.Millisecond)
			counter = current + 1
		}()
	}

	wg.Wait()

	if counter != goroutines {
		b.Fatalf("expected counter to be %d, got %d", goroutines, counter)
	}
}

func BenchmarkShardedMutexExp_Concurrent(b *testing.B) {
	const goroutines = 100
	var mu *ShardedMutex = NewShardedMutex(goroutines)
	counter := 0
	wg := sync.WaitGroup{}

	wg.Add(goroutines)
	b.ResetTimer()

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			key := CalculateKey(goroutines)
			mu.Lock(key)

			current := counter
			time.Sleep(10 * time.Millisecond)
			counter = current + 1
			mu.Unlock(key)
		}()
	}

	wg.Wait()

	if counter != goroutines {
		b.Fatalf("expected counter to be %d, got %d", goroutines, counter)
	}
}
