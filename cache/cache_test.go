package cache

import (
    "testing"
    "trabalho_pratico/main_memory"
    "fmt"
)



func TestNewCacheBlock(t *testing.T) {
    cacheSize := 40
    address := 0
    block := &CacheBlock{
        address: address,
        data:    make([]string, int(float64(cacheSize)*0.2)),
        state:   Invalid,
        time:    0,
    }

    if block == nil {
        t.Fatalf("Expected NewCacheBlock to return a valid block, got nil")
    }

    if block.address != address {
        t.Errorf("Expected address %d, got %d", address, block.address)
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

    // Verifica se a lineSize está correta (20% do tamanho da cache)
    expectedLineSize := int(float64(expectedCacheSize) * 0.2)
    if cache.lineSize != expectedLineSize {
        t.Errorf("Expected line size %d, got %d", expectedLineSize, cache.lineSize)
    }

    // Verifica se o mapa de blocos foi inicializado
    if cache.blocks == nil {
        t.Error("Expected blocks map to be initialized, got nil")
    }

    // Verifica se o contador FIFO foi inicializado
    if cache.nextTime != 0 {
        t.Errorf("Expected nextTime to be 0, got %d", cache.nextTime)
    }

    // Verifica se não há blocos inicialmente
    if len(cache.blocks) != 0 {
        t.Errorf("Expected empty blocks map, got %d blocks", len(cache.blocks))
    }
}

func TestCacheInitialization(t *testing.T) {
    mainMemorySize := 1000
    cache := NewCache(mainMemorySize)
    expectedCacheSize := int(float64(mainMemorySize) * 0.4)

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
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.got != tt.expected {
                t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.got)
            }
        })
    }

    // Verifica se o mapa de blocos foi inicializado corretamente
    if cache.blocks == nil {
        t.Error("Blocks map was not initialized")
    }
}

func TestCacheRead(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    cache := NewCache(mainMemory.Size)

    // Inicializa a memória principal
    testValue := "valor teste"
    testAddress := 42
    err := mainMemory.Write(testAddress, testValue)
    if err != nil {
        t.Fatalf("Erro ao inicializar memória principal: %v", err)
    }

    // Teste 1: Cache miss - primeira leitura
    value := cache.read(testAddress, mainMemory)
    if value != testValue {
        t.Errorf("Valor incorreto no cache miss. Esperado: %q, Recebido: %q", testValue, value)
    }

    // Verifica se o bloco foi adicionado corretamente
    found := false
    for _, block := range cache.blocks {
        if block.address == testAddress {
            found = true
            if block.data[0] != testValue {
                t.Errorf("Dado incorreto no bloco. Esperado: %q, Recebido: %q", testValue, block.data[0])
            }
            if block.state != Exclusive {
                t.Errorf("Estado incorreto após primeira leitura. Esperado: Exclusive, Recebido: %v", block.state)
            }
            break
        }
    }
    if !found {
        t.Error("Bloco não foi adicionado à cache após miss")
    }

    // Teste 2: Cache hit - segunda leitura
    value = cache.read(testAddress, mainMemory)
    if value != testValue {
        t.Errorf("Valor incorreto no cache hit. Esperado: %q, Recebido: %q", testValue, value)
    }
}

func TestCacheWrite(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    cache := NewCache(mainMemory.Size)

    testAddress := 42
    testValue := "valor teste"

    // Teste 1: Primeira escrita
    err := cache.Write(testAddress, testValue, mainMemory)
    if err != nil {
        t.Fatalf("Erro na escrita: %v", err)
    }

    // Verifica se o bloco foi adicionado corretamente
    found := false
    for _, block := range cache.blocks {
        if block.address == testAddress {
            found = true
            if block.data[0] != testValue {
                t.Errorf("Dado incorreto no bloco. Esperado: %q, Recebido: %q", testValue, block.data[0])
            }
            if block.state != Modified {
                t.Errorf("Estado incorreto após escrita. Esperado: Modified, Recebido: %v", block.state)
            }
            break
        }
    }
    if !found {
        t.Error("Bloco não foi adicionado à cache após escrita")
    }

    // Teste 2: Sobrescrita no mesmo endereço
    newValue := "novo valor"
    err = cache.Write(testAddress, newValue, mainMemory)
    if err != nil {
        t.Fatalf("Erro na sobrescrita: %v", err)
    }

    // Verifica se o valor foi atualizado
    found = false
    for _, block := range cache.blocks {
        if block.address == testAddress {
            found = true
            if block.data[0] != newValue {
                t.Errorf("Dado incorreto após sobrescrita. Esperado: %q, Recebido: %q", newValue, block.data[0])
            }
            if block.state != Modified {
                t.Errorf("Estado incorreto após sobrescrita. Esperado: Modified, Recebido: %v", block.state)
            }
            break
        }
    }
    if !found {
        t.Error("Bloco não foi encontrado após sobrescrita")
    }
}

func TestCacheFIFOReplacement(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    cache := NewCache(10) // Cache pequena para forçar substituições

    // Preenche a cache até ficar cheia
    for i := 0; i < cache.size; i++ {
        value := fmt.Sprintf("valor %d", i)
        err := cache.Write(i, value, mainMemory)
        if err != nil {
            t.Fatalf("Erro ao preencher cache: %v", err)
        }
    }

    // Tenta adicionar um novo bloco para forçar substituição FIFO
    newAddress := cache.size
    newValue := "novo valor"
    err := cache.Write(newAddress, newValue, mainMemory)
    if err != nil {
        t.Fatalf("Erro na substituição FIFO: %v", err)
    }

    // Verifica se o bloco mais antigo foi substituído
    oldestValue := "valor 0"
    found := false
    for _, block := range cache.blocks {
        if block.data[0] == oldestValue {
            found = true
            break
        }
    }
    if found {
        t.Error("O bloco mais antigo não foi substituído pelo FIFO")
    }

    // Verifica se o novo bloco está presente
    found = false
    for _, block := range cache.blocks {
        if block.address == newAddress && block.data[0] == newValue {
            found = true
            break
        }
    }
    if !found {
        t.Error("Novo bloco não foi adicionado corretamente após substituição FIFO")
    }
}

func TestGetDisplayBlocks(t *testing.T) {
    mainMemory := main_memory.NewMemory(1000)
    cache := NewCache(mainMemory.Size)

    // Adiciona alguns blocos para teste
    testAddresses := []int{0, 1, 2}
    for _, addr := range testAddresses {
        value := fmt.Sprintf("valor %d", addr)
        err := cache.Write(addr, value, mainMemory)
        if err != nil {
            t.Fatalf("Erro ao preparar cache para display: %v", err)
        }
    }

    blocks := cache.GetDisplayBlocks()
    if len(blocks) == 0 {
        t.Error("GetDisplayBlocks retornou array vazio")
    }

    // Verifica se cada página contém informações corretas
    for _, page := range blocks {
        if len(page) == 0 {
            t.Error("Página vazia encontrada")
        }
        for _, line := range page {
            if line == "" {
                t.Error("Linha vazia encontrada")
            }
        }
    }
}