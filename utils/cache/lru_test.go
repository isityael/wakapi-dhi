package cache

import (
	"runtime"
	"testing"
)

func TestLRUCacheEvictsLeastRecentlyUsedEntry(t *testing.T) {
	cache := NewLRU[string, int](2)

	cache.Add("oldest", 1)
	cache.Add("newest", 2)
	if got, ok := cache.Get("oldest"); !ok || got != 1 {
		t.Fatalf("expected oldest entry to exist before eviction, got %v, %v", got, ok)
	}

	cache.Add("third", 3)

	if _, ok := cache.Get("newest"); ok {
		t.Fatal("expected least recently used entry to be evicted")
	}
	if got, ok := cache.Get("oldest"); !ok || got != 1 {
		t.Fatalf("expected recently used entry to remain, got %v, %v", got, ok)
	}
	if got, ok := cache.Get("third"); !ok || got != 3 {
		t.Fatalf("expected newly added entry to remain, got %v, %v", got, ok)
	}
}

func TestLRUCacheRemoveAndContains(t *testing.T) {
	cache := NewLRU[string, bool](1)

	cache.Add("key", true)
	if !cache.Contains("key") {
		t.Fatal("expected cache to contain added key")
	}

	cache.Remove("key")
	if cache.Contains("key") {
		t.Fatal("expected cache not to contain removed key")
	}
	if _, ok := cache.Get("key"); ok {
		t.Fatal("expected removed key lookup to miss")
	}
}

func TestLRUCacheDoesNotPreallocateCapacity(t *testing.T) {
	// fs.NewExistsFS uses a 1<<24 capacity; preallocating it cost ~850 MiB at
	// startup and OOM-killed the pod under a 1 GiB limit.
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	cache := NewLRU[string, bool](1 << 24)
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(cache)

	if grown := after.TotalAlloc - before.TotalAlloc; grown > 1<<20 {
		t.Fatalf("expected constructing a large-capacity cache to stay under 1 MiB, allocated %d bytes", grown)
	}
}
