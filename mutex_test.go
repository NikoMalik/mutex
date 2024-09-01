package mutex_test

import (
	"sync"
	"testing"
	"time"

	"github.com/NikoMalik/mutex"
)

func TestMutexExp_Concurrent(t *testing.T) {
	var mu mutex.MutexExp
	counter := 0
	wg := sync.WaitGroup{}

	const goroutines = 100

	wg.Add(goroutines)

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
		t.Fatalf("expected counter to be %d, got %d", goroutines, counter)
	}
}

func TestMutexExp_PanicOnUnlock(t *testing.T) {
	var mu mutex.MutexExp
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on unlocking unlocked mutex, but there was no panic")
		}
	}()

	mu.Unlock()
}
