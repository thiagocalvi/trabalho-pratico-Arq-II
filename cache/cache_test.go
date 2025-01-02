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
    cache := NewCache(4)

    // Inicializa a memória principal
    for i := 0; i < 10; i++ {
        _ = mainMemory.Write(i, fmt.Sprintf("valor %d", i))
    }

    // Testa leitura com cache miss
    value := cache.read(1, mainMemory)
    if value != "valor 1" {
        t.Errorf("Erro na leitura (miss). Esperado: %q, Recebido: %q", "valor 1", value)
    }

    // Verifica se o bloco foi adicionado corretamente
    tag := 1 % cache.size
    block, exists := cache.blocks[tag]
    if !exists || len(block.data) == 0 || block.data[0] != "valor 1" {
        t.Errorf("Bloco não foi adicionado corretamente na cache após miss.")
    }

    // Testa leitura com cache hit
    value = cache.read(1, mainMemory)
    if value != "valor 1" {
        t.Errorf("Erro na leitura (hit). Esperado: %q, Recebido: %q", "valor 1", value)
    }
}

func TestReplaceBlock(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    cache := NewCache(mainMemory.Size)

    // Simula o preenchimento total da cache
    for i := 0; i < cache.size; i++ {
        cache.queue = append(cache.queue, i)
        cache.blocks[i] = &CacheBlock{
            tag:   i,
            data:  []string{fmt.Sprintf("valor antigo %d", i)},
            state: Modified,
        }
    }

    // Escreve um novo valor, forçando a substituição
    address := 100
    value := "novo valor"
    cache.ReplaceBlock(address, value, mainMemory)

    // Verifica se o bloco mais antigo foi removido
    oldest := 0
    if _, exists := cache.blocks[oldest]; exists {
        t.Errorf("Bloco mais antigo (%d) não foi removido corretamente", oldest)
    }

    // Verifica se o valor antigo foi sincronizado com a memória principal
    if memValue, _ := mainMemory.Read(oldest); memValue != "valor antigo 0" {
        t.Errorf("Valor antigo não foi escrito na memória principal. Esperado: %q, Recebido: %q", "valor antigo 0", memValue)
    }

    // Verifica o novo bloco
    newTag := address % cache.size
    newBlock, exists := cache.blocks[newTag]
    if !exists || newBlock.data[0] != value || newBlock.state != Modified {
        t.Errorf("Novo bloco não foi atualizado corretamente. Esperado: {Tag: %d, Dados: %q, Estado: Modified}", newTag, value)
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
