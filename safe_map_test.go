package test_task

import (
	"math/rand"
	"sync"
	"testing"
)

func buildOps(year int, times int, seed int64) []int {
	ops := make([]int, 0, year*times)
	for key := 1; key <= year; key++ {
		for range times {
			ops = append(ops, key)
		}
	}
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(ops), func(i, j int) { ops[i], ops[j] = ops[j], ops[i] })
	return ops
}

func worker(m *SafeMap, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for key := range jobs {
		value, exists := m.Get(key)
		if !exists {
			value = 0
		}
		m.Set(key, value+1)
	}
}

func TestSafeMapConcurrentAccess(t *testing.T) {

	// В качестве года выбран 1968 — в этом году Боб Бимон установил мировой рекорд в прыжке в длину.
	// 1968 кратно 4, поэтому ключи удобно шардировать между 4 воркерами.
	const year = 1968

	m := NewSafeMap()
	ops := buildOps(year, 3, 1968)

	var wg sync.WaitGroup
	workerJobs := [4]chan int{
		make(chan int, 64),
		make(chan int, 64),
		make(chan int, 64),
		make(chan int, 64),
	}

	// Четыре воркера потребляют ключи параллельно (fan-out).
	for i := range 4 {
		wg.Add(1)
		go worker(m, workerJobs[i], &wg)
	}

	// Шардируем ключи синхронно, чтобы не нарушать условие про 4 горутины.
	for _, key := range ops {
		workerJobs[(key-1)%4] <- key
	}
	for i := range 4 {
		close(workerJobs[i])
	}

	wg.Wait()

	snapshot := m.Snapshot()

	// для отладки
	// t.Logf("map size=%d", len(snapshot))
	// for k := 1; k <= year; k++ {
	// 	t.Logf("key=%d value=%d", k, snapshot[k])
	// }
	if len(snapshot) != year {
		t.Fatalf("unexpected map size: got %d, want %d", len(snapshot), year)
	}

	for key := 1; key <= year; key++ {
		value, ok := snapshot[key]
		if !ok {
			t.Fatalf("key %d not found", key)
		}
		if value != 3 {
			t.Fatalf("unexpected value for key %d: got %d, want 3", key, value)
		}
	}

	accessCount, insertCount := m.Stats()
	if accessCount != year*3 {
		t.Fatalf("unexpected accessCount: got %d, want %d", accessCount, year*3)
	}
	if insertCount != year {
		t.Fatalf("unexpected insertCount: got %d, want %d", insertCount, year)
	}
}
