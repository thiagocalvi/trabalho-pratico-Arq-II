package cache

import "testing"

// Criação de uma nova Cache com tamanho 5
func TestNewCache(t *testing.T) {
	mainMemorySize := 100
	cache := NewCache(mainMemorySize)

	// Verifica se a cache foi inicializada corretamente
	expectedSize := int(float64(mainMemorySize) * 0.4)

	if cache.size != expectedSize {
		t.Errorf("esperado tamanho da cache %d, mas obteve %d", expectedSize, cache.size)
	}

	if len(cache.blocks) != 0 {
		t.Errorf("esperado 0 blocos na cache, mas obteve %d", len(cache.blocks))
	}

	if len(cache.queue) != 0 {
		t.Errorf("esperado fila vazia, mas obteve %d elementos", len(cache.queue))
	}

}
