package cache

import "testing"

func TestNewCacheBlock(t *testing.T) {
	cacheSize := 40
	tagNumber := 0
	block := NewCacheBlock(cacheSize, tagNumber)

	if block == nil {
		t.Fatalf("Expected NewCacheBlock to return a valid block, got nil")
	}

	if block.tag != tagNumber {
		t.Errorf("Expected tag %d, got %d", tagNumber, block.tag)
	}

	expectedLineSize := int(float64(cacheSize) * 0.2)
	if cap(block.data) != expectedLineSize {
		t.Errorf("Expected data capacity %d, got %d", expectedLineSize, cap(block.data))
	}

	if block.state != Invalid {
		t.Errorf("Expected initial state Invalid, got %v", block.state)
	}
}

func TestNewCache(t *testing.T) {
	mainMemorySize := 1000
	cache := NewCache(mainMemorySize)

	if cache == nil {
		t.Fatalf("Expected NewCache to return a valid cache, got nil")
	}

	expectedCacheSize := int(float64(mainMemorySize) * 0.4)
	if cache.size != expectedCacheSize {
		t.Errorf("Expected cache size %d, got %d", expectedCacheSize, cache.size)
	}

	if len(cache.blocks) != expectedCacheSize {
		t.Errorf("Expected %d blocks, got %d", expectedCacheSize, len(cache.blocks))
	}

	for i := 0; i < expectedCacheSize; i++ {
		block, exists := cache.blocks[i]
		if !exists {
			t.Errorf("Expected block %d to exist in cache.blocks", i)
		}
		if block.tag != i {
			t.Errorf("Expected block %d to have tag %d, got %d", i, i, block.tag)
		}
		if block.state != Invalid {
			t.Errorf("Expected block %d to have state Invalid, got %v", i, block.state)
		}
	}
}

func TestCacheQueueInitialization(t *testing.T) {
	mainMemorySize := 1000
	cache := NewCache(mainMemorySize)

	expectedCacheSize := int(float64(mainMemorySize) * 0.4)
	if len(cache.queue) != 0 {
		t.Errorf("Expected queue to be empty initially, got length %d", len(cache.queue))
	}

	if cap(cache.queue) != expectedCacheSize {
		t.Errorf("Expected queue capacity %d, got %d", expectedCacheSize, cap(cache.queue))
	}
}
