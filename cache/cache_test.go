package cache

import (
	"testing"
	"trabalho_pratico/main_memory"
	"fmt"
)

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

    // Verifica se a cache foi criada corretamente
    expectedCacheSize := int(float64(mainMemorySize) * 0.4)
    if cache.size != expectedCacheSize {
        t.Errorf("Expected cache size %d, got %d", expectedCacheSize, cache.size)
    }

    // Verifica o número de blocos
    if len(cache.blocks) != expectedCacheSize {
        t.Errorf("Expected %d blocks, got %d", expectedCacheSize, len(cache.blocks))
    }

    // Verifica cada bloco individualmente
    for i := 0; i < expectedCacheSize; i++ {
        block, exists := cache.blocks[i]
        if !exists {
            t.Errorf("Expected block %d to exist", i)
            continue
        }

        if block.tag != i {
            t.Errorf("Expected block %d to have tag %d, got %d", i, i, block.tag)
        }

        if block.state != Invalid {
            t.Errorf("Expected block %d to have state Invalid, got %v", i, block.state)
        }
    }

    // Verifica se a fila está vazia inicialmente
    if len(cache.queue) != 0 {
        t.Errorf("Expected queue to be empty, got length %d", len(cache.queue))
    }
}

func TestCacheQueueInitialization(t *testing.T) {
    mainMemorySize := 1000
    cache := NewCache(mainMemorySize)

    expectedCacheSize := int(float64(mainMemorySize) * 0.4)

    // Verifica que a fila está vazia
    if len(cache.queue) != 0 {
        t.Errorf("Expected queue to be empty initially, got length %d", len(cache.queue))
    }

    // Verifica a capacidade da fila
    if cap(cache.queue) != expectedCacheSize {
        t.Errorf("Expected queue capacity %d, got %d", expectedCacheSize, cap(cache.queue))
    }
}

func TestCacheWrite(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    cache := NewCache(mainMemory.Size)

    address := 10
    value := "teste"

    // Realiza a escrita na cache
    err := cache.Write(address, value, mainMemory)
    if err != nil {
        t.Errorf("Erro ao escrever na cache: %v", err)
    }

    // Verifica o bloco correspondente
    tag := address % cache.size
    block, exists := cache.blocks[tag]
    if !exists {
        t.Fatalf("Bloco da tag %d não foi criado", tag)
    }

    if block.state != Modified {
        t.Errorf("Estado incorreto do bloco. Esperado: Modified, Recebido: %v", block.state)
    }

    if len(block.data) == 0 || block.data[0] != value {
        t.Errorf("Valor incorreto no bloco. Esperado: %q, Recebido: %q", value, block.data[0])
    }

    // Verifica se a tag foi adicionada à fila
    if len(cache.queue) == 0 || cache.queue[len(cache.queue)-1] != tag {
        t.Errorf("Tag %d não foi adicionada corretamente à fila", tag)
    }
}

func TestCacheRead(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    cache := NewCache(mainMemory.Size)

    // Caso 1: Cache miss
    addressMiss := 10
    valueMiss := "miss value"
    _ = mainMemory.Write(addressMiss, valueMiss)

    result, err := cache.Read(addressMiss, mainMemory)
    if err != nil {
        t.Errorf("Erro ao ler da cache (miss): %v", err)
    }
    if result != valueMiss {
        t.Errorf("Valor incorreto no miss. Esperado: %q, Recebido: %q", valueMiss, result)
    }

    // Verifica se o bloco foi atualizado
    tagMiss := addressMiss % cache.size
    blockMiss, existsMiss := cache.blocks[tagMiss]
    if !existsMiss || blockMiss.state != Shared || blockMiss.data[0] != valueMiss {
        t.Errorf("Bloco não atualizado corretamente no miss. Esperado: %q, Estado: Shared", valueMiss)
    }

    // Caso 2: Cache hit
    addressHit := addressMiss
    result, err = cache.Read(addressHit, mainMemory)
    if err != nil {
        t.Errorf("Erro ao ler da cache (hit): %v", err)
    }
    if result != valueMiss {
        t.Errorf("Valor incorreto no hit. Esperado: %q, Recebido: %q", valueMiss, result)
    }
}

func TestReplaceBlock(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    cache := NewCache(3) // Cache com tamanho limitado a 3 blocos

    // Preenche a cache com 3 blocos
    for i := 0; i < 3; i++ {
        data := fmt.Sprintf("data-%d", i)
        cache.ReplaceBlock(i, data, mainMemory)
    }

    // Garante que os 3 blocos iniciais estão na cache
    for i := 0; i < 3; i++ {
        if block, exists := cache.blocks[i]; !exists || block.data[0] != fmt.Sprintf("data-%d", i) {
            t.Errorf("Bloco %d não encontrado ou incorreto", i)
        }
    }

    // Substitui um bloco (deve remover o mais antigo)
    newTag := 3
    newData := "new-data"
    cache.ReplaceBlock(newTag, newData, mainMemory)

    // Verifica se o bloco mais antigo foi removido
    if _, exists := cache.blocks[0]; exists {
        t.Errorf("Bloco mais antigo (tag 0) não foi removido corretamente")
    }

    // Verifica se o novo bloco foi adicionado
    if block, exists := cache.blocks[newTag]; !exists || block.data[0] != newData {
        t.Errorf("Novo bloco não foi adicionado corretamente. Esperado: %q, Recebido: %v", newData, block)
    }

    // Verifica se os outros blocos ainda estão na cache
    for i := 1; i < 3; i++ {
        if block, exists := cache.blocks[i]; !exists || block.data[0] != fmt.Sprintf("data-%d", i) {
            t.Errorf("Bloco %d foi removido incorretamente", i)
        }
    }
}

func TestCacheGetDisplayBlocks(t *testing.T) {
	mainMemory := main_memory.NewMemory(1000)
	cache := NewCache(mainMemory.Size)

	for i := 0; i < cache.size; i++ {
		cache.blocks[i] = &CacheBlock{
			tag:   i,
			data:  []string{fmt.Sprintf("value %d", i)},
			state: Shared,
		}
	}

	blocks := cache.GetDisplayBlocks()
	expectedBlockCount := 5
	if len(blocks) != expectedBlockCount {
		t.Errorf("Incorrect number of display blocks. Expected: %d, Got: %d", expectedBlockCount, len(blocks))
	}

	expectedBlockSize := cache.size / expectedBlockCount
	for i, block := range blocks {
		if len(block) != expectedBlockSize {
			t.Errorf("Incorrect size for block %d. Expected: %d, Got: %d", i+1, expectedBlockSize, len(block))
		}

		start := i * expectedBlockSize
		for j, line := range block {
			expectedLine := fmt.Sprintf("Tag %d, State: %v, Data: [value %d]", start+j, Shared, start+j)
			if line != expectedLine {
				t.Errorf("Incorrect content in block %d, line %d. Expected: %q, Got: %q", i+1, j+1, expectedLine, line)
			}
		}
	}
}

func TestCacheDisplay(t *testing.T) {
	mainMemory := main_memory.NewMemory(1000)
	cache := NewCache(mainMemory.Size)

	for i := 0; i < cache.size; i++ {
		cache.blocks[i] = &CacheBlock{
			tag:   i,
			data:  []string{fmt.Sprintf("value %d", i)},
			state: Shared,
		}
	}

	// Captura a saída do Display para validação
	// TODO: Implementar captura de saída usando bytes.Buffer
	cache.Display()
}
