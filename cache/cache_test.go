package cache

import (
    "testing"
    "trabalho_pratico/main_memory"
)

func TestCacheInitialization(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    bus := NewCoherencyBus(mainMemory)
    cache := NewCache(mainMemory.Size, bus)
    expectedCacheSize := int(float64(mainMemory.Size) * 0.4)

    // Testes básicos de inicialização
    tests := []struct {
        name     string
        got      interface{}
        expected interface{}
    }{
        {"Cache Size", cache.size, expectedCacheSize},
        {"Line Size", cache.lineSize, int(float64(expectedCacheSize) * 0.2)},
        {"Initial Blocks Length", len(cache.blocks), 0},
        {"Next Time Counter", cache.nextTime, 0},
        {"Cache ID", cache.id, 0}, // Primeira cache adicionada ao bus
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.got != tt.expected {
                t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.got)
            }
        })
    }

    if cache.blocks == nil {
        t.Error("Blocks map was not initialized")
    }

    if cache.bus != bus {
        t.Error("Bus reference not properly set")
    }
}

func TestCacheRead(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    bus := NewCoherencyBus(mainMemory)
    cache := NewCache(mainMemory.Size, bus)

    // Inicializa a memória principal
    testValue := "valor teste"
    testAddress := 42
    err := mainMemory.Write(testAddress, testValue)
    if err != nil {
        t.Fatalf("Erro ao inicializar memória principal: %v", err)
    }

    // Teste 1: Cache miss - primeira leitura
    value, err := cache.Read(testAddress)
    if err != nil {
        t.Fatalf("Erro na primeira leitura: %v", err)
    }
    if value != testValue {
        t.Errorf("Valor incorreto no cache miss. Esperado: %q, Recebido: %q", testValue, value)
    }

    // Verifica se o bloco foi adicionado corretamente
    block, exists := cache.blocks[testAddress]
    if !exists {
        t.Error("Bloco não foi adicionado à cache após miss")
    } else {
        if block.data[0] != testValue {
            t.Errorf("Dado incorreto no bloco. Esperado: %q, Recebido: %q", testValue, block.data[0])
        }
        if block.state != Exclusive {
            t.Errorf("Estado incorreto após primeira leitura. Esperado: Exclusive, Recebido: %v", block.state)
        }
    }

    // Teste 2: Cache hit - segunda leitura
    value, err = cache.Read(testAddress)
    if err != nil {
        t.Fatalf("Erro na segunda leitura: %v", err)
    }
    if value != testValue {
        t.Errorf("Valor incorreto no cache hit. Esperado: %q, Recebido: %q", testValue, value)
    }
}

func TestCacheWrite(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    bus := NewCoherencyBus(mainMemory)
    cache := NewCache(mainMemory.Size, bus)

    testAddress := 42
    testValue := "valor teste"

    // Teste 1: Primeira escrita
    err := cache.Write(testAddress, testValue)
    if err != nil {
        t.Fatalf("Erro na escrita: %v", err)
    }

    // Verifica se o bloco foi adicionado corretamente
    block, exists := cache.blocks[testAddress]
    if !exists {
        t.Error("Bloco não foi adicionado à cache após escrita")
    } else {
        if block.data[0] != testValue {
            t.Errorf("Dado incorreto no bloco. Esperado: %q, Recebido: %q", testValue, block.data[0])
        }
        if block.state != Modified {
            t.Errorf("Estado incorreto após escrita. Esperado: Modified, Recebido: %v", block.state)
        }
    }

    // Teste 2: Sobrescrita no mesmo endereço
    newValue := "novo valor"
    err = cache.Write(testAddress, newValue)
    if err != nil {
        t.Fatalf("Erro na sobrescrita: %v", err)
    }

    block, exists = cache.blocks[testAddress]
    if !exists {
        t.Error("Bloco não foi encontrado após sobrescrita")
    } else {
        if block.data[0] != newValue {
            t.Errorf("Dado incorreto após sobrescrita. Esperado: %q, Recebido: %q", newValue, block.data[0])
        }
        if block.state != Modified {
            t.Errorf("Estado incorreto após sobrescrita. Esperado: Modified, Recebido: %v", block.state)
        }
    }
}

func TestMultipleCachesCoherence(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    bus := NewCoherencyBus(mainMemory)
    cache1 := NewCache(mainMemory.Size, bus)
    cache2 := NewCache(mainMemory.Size, bus)

    testAddress := 42
    testValue := "valor teste"

    // Cache 1 escreve o valor
    err := cache1.Write(testAddress, testValue)
    if err != nil {
        t.Fatalf("Erro na escrita inicial: %v", err)
    }

    // Cache 2 lê o valor
    value, err := cache2.Read(testAddress)
    if err != nil {
        t.Fatalf("Erro na leitura da segunda cache: %v", err)
    }
    if value != testValue {
        t.Errorf("Valor incorreto na segunda cache. Esperado: %q, Recebido: %q", testValue, value)
    }

    // Verifica estados
    block1 := cache1.blocks[testAddress]
    block2 := cache2.blocks[testAddress]

    // Na implementação atual, o bloco permanece Modified na cache1 e Shared na cache2
    if block1.state != Modified {
        t.Errorf("Estado incorreto na cache 1. Esperado: Modified, Recebido: %v", block1.state)
    }
    if block2.state != Shared {
        t.Errorf("Estado incorreto na cache 2. Esperado: Shared, Recebido: %v", block2.state)
    }
}

func TestCacheFIFOReplacement(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    bus := NewCoherencyBus(mainMemory)
    
    // Criar cache com estrutura modificada para teste
    cache := &Cache{
        size:     3, // Tamanho fixo para teste
        lineSize: 1,
        blocks:   make(map[int]*CacheBlock),
        nextTime: 0,
        bus:      bus,
    }
    bus.AddCache(cache)

    // Preenche a cache até ficar cheia
    for i := 0; i < 3; i++ {
        err := cache.Write(i, string(rune('A'+i)))
        if err != nil {
            t.Fatalf("Erro ao preencher cache: %v", err)
        }
    }

    initialSize := len(cache.blocks)
    if initialSize != 3 {
        t.Errorf("Cache deveria estar cheia com 3 blocos, tem %d", initialSize)
    }

    // Adiciona um novo bloco para forçar substituição FIFO
    newAddress := 3
    newValue := "X"
    err := cache.Write(newAddress, newValue)
    if err != nil {
        t.Fatalf("Erro na substituição FIFO: %v", err)
    }

    // Verifica se o tamanho da cache permanece o mesmo
    if len(cache.blocks) != 3 {
        t.Errorf("Tamanho da cache deveria permanecer 3, está %d", len(cache.blocks))
    }

    // Verifica se o novo bloco está presente
    newBlock, exists := cache.blocks[newAddress]
    if !exists {
        t.Error("Novo bloco não foi adicionado após substituição FIFO")
    } else if newBlock.data[0] != newValue {
        t.Errorf("Dado incorreto no novo bloco. Esperado: %q, Recebido: %q", newValue, newBlock.data[0])
    }

    // Verifica se pelo menos um dos blocos antigos foi removido
    blocksRemoved := 0
    for i := 0; i < 3; i++ {
        if _, exists := cache.blocks[i]; !exists {
            blocksRemoved++
        }
    }
    if blocksRemoved == 0 {
        t.Error("Nenhum bloco antigo foi removido após a substituição FIFO")
    }
}