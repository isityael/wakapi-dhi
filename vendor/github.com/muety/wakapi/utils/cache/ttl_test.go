package cache

import (
	"testing"
	"time"
)

func TestCacheDefaultExpirationExpiresItems(t *testing.T) {
	cache := New(20*time.Millisecond, time.Hour)

	cache.SetDefault("key", "value")
	if got, ok := cache.Get("key"); !ok || got != "value" {
		t.Fatalf("expected cached value before expiration, got %v, %v", got, ok)
	}

	time.Sleep(30 * time.Millisecond)
	if _, ok := cache.Get("key"); ok {
		t.Fatal("expected default-expiration item to expire")
	}
}

func TestCacheNoExpirationKeepsItems(t *testing.T) {
	cache := New(10*time.Millisecond, time.Hour)

	cache.Set("key", "value", NoExpiration)
	time.Sleep(20 * time.Millisecond)

	if got, ok := cache.Get("key"); !ok || got != "value" {
		t.Fatalf("expected no-expiration item to remain, got %v, %v", got, ok)
	}
}

func TestCacheDeleteFlushAndItemCount(t *testing.T) {
	cache := New(NoExpiration, time.Hour)

	cache.SetDefault("one", 1)
	cache.SetDefault("two", 2)
	cache.Delete("one")

	if _, ok := cache.Get("one"); ok {
		t.Fatal("expected deleted item to miss")
	}
	if count := cache.ItemCount(); count != 1 {
		t.Fatalf("expected one item after delete, got %d", count)
	}

	cache.Flush()
	if count := cache.ItemCount(); count != 0 {
		t.Fatalf("expected empty cache after flush, got %d", count)
	}
}

func TestCacheItemsOmitsExpiredEntries(t *testing.T) {
	cache := New(10*time.Millisecond, time.Hour)

	cache.SetDefault("expired", "value")
	cache.Set("kept", "value", NoExpiration)
	time.Sleep(20 * time.Millisecond)

	items := cache.Items()
	if _, ok := items["expired"]; ok {
		t.Fatal("expected expired item to be omitted")
	}
	if item, ok := items["kept"]; !ok || item.Object != "value" {
		t.Fatalf("expected kept item snapshot, got %v, %v", item, ok)
	}
}

func TestCacheIncrementIntTypes(t *testing.T) {
	cache := New(NoExpiration, time.Hour)

	cache.SetDefault("int", 1)
	if got, err := cache.IncrementInt("int", 2); err != nil || got != 3 {
		t.Fatalf("expected int increment to return 3, got %d, %v", got, err)
	}

	cache.SetDefault("int64", int64(4))
	if got, err := cache.IncrementInt64("int64", 3); err != nil || got != 7 {
		t.Fatalf("expected int64 increment to return 7, got %d, %v", got, err)
	}

	if _, err := cache.IncrementInt("missing", 1); err == nil {
		t.Fatal("expected missing increment to fail")
	}
}
