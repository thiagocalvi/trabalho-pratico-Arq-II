package cache

import (
    "fmt"
    "sync"
    "trabalho_pratico/main_memory"
)

// Estados MESIF
type CacheBlockState uint8

const (
    Modified CacheBlockState = iota
    Exclusive
    Shared
    Invalid
    Forward
)

func (state CacheBlockState) String() string {
    switch state {
    case Modified:
        return "Modified"
    case Exclusive:
        return "Exclusive"
    case Shared:
        return "Shared"
    case Invalid:
        return "Invalid"
    case Forward:
        return "Forward"
    default:
        return "Unknown"
    }
}

// Bus de coerência para comunicação entre caches
type CoherencyBus struct {
    mu      sync.RWMutex
    caches  []*Cache
    memory  *main_memory.MainMemory
}

func NewCoherencyBus(memory *main_memory.MainMemory) *CoherencyBus {
    return &CoherencyBus{
        memory: memory,
        caches: make([]*Cache, 0),
    }
}

// Adiciona uma cache ao bus
func (b *CoherencyBus) AddCache(cache *Cache) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.caches = append(b.caches, cache)
    cache.id = len(b.caches) - 1
}

type CacheBlock struct {
    address int             
    data    []string        
    state   CacheBlockState 
    time    int            
}

type Cache struct {
    id       int                 // ID único da cache
    size     int                 // Número de linhas
    lineSize int                 // Tamanho da linha
    blocks   map[int]*CacheBlock // Blocos da cache
    nextTime int                 // Contador FIFO
    bus      *CoherencyBus      // Referência ao bus
    mu       sync.RWMutex       // Mutex para operações concorrentes
}

func NewCache(mainMemorySize int, bus *CoherencyBus) *Cache {
    cacheSize := int(float64(mainMemorySize) * 0.4)
    lineSize := int(float64(cacheSize) * 0.2)

    cache := &Cache{
        size:     cacheSize,
        lineSize: lineSize,
        blocks:   make(map[int]*CacheBlock),
        nextTime: 0,
        bus:      bus,
    }
    
    bus.AddCache(cache)
    return cache
}

// Busca um bloco em outras caches
func (c *Cache) findInOtherCaches(address int) (*Cache, *CacheBlock) {
    c.bus.mu.RLock()
    defer c.bus.mu.RUnlock()

    for _, otherCache := range c.bus.caches {
        if otherCache.id == c.id {
            continue
        }

        otherCache.mu.RLock()
        if block, exists := otherCache.blocks[address]; exists && 
           (block.state == Modified || block.state == Forward) {
            otherCache.mu.RUnlock()
            return otherCache, block
        }
        otherCache.mu.RUnlock()
    }
    return nil, nil
}

// Invalida cópias em outras caches
func (c *Cache) invalidateOtherCopies(address int) {
    c.bus.mu.RLock()
    defer c.bus.mu.RUnlock()

    for _, otherCache := range c.bus.caches {
        if otherCache.id == c.id {
            continue
        }

        otherCache.mu.Lock()
        if block, exists := otherCache.blocks[address]; exists {
            block.state = Invalid
            fmt.Printf("Cache %d: Invalidando bloco %d na cache %d\n",
                c.id, address, otherCache.id)
        }
        otherCache.mu.Unlock()
    }
}

func (c *Cache) Read(address int) (string, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    fmt.Printf("Cache %d: Lendo endereço %d\n", c.id, address)

    // 1. Verifica cache local
    if block, exists := c.blocks[address]; exists {
        switch block.state {
        case Modified, Exclusive, Forward:
            fmt.Printf("Cache %d: Hit - Estado %s\n", c.id, block.state)
            return block.data[0], nil
        case Shared:
            fmt.Printf("Cache %d: Hit - Estado Shared\n", c.id)
            return block.data[0], nil
        case Invalid:
            fmt.Printf("Cache %d: Bloco inválido, buscando dados atualizados\n", c.id)
        }
    }

    // 2. Cache miss - procura em outras caches
    sourceCache, sourceBlock := c.findInOtherCaches(address)
    if sourceCache != nil {
        fmt.Printf("Cache %d: Obtendo dados da cache %d\n", c.id, sourceCache.id)
        
        // Cria novo bloco local
        newBlock := &CacheBlock{
            address: address,
            data:    []string{sourceBlock.data[0]},
            state:   Shared,
            time:    c.nextTime,
        }
        c.nextTime++

        // Atualiza estado da cache fonte
        sourceCache.mu.Lock()
        if sourceBlock.state == Exclusive {
            sourceBlock.state = Forward
        }
        sourceCache.mu.Unlock()

        // Armazena localmente
        if len(c.blocks) >= c.size {
            c.replaceFIFO(address, newBlock.data[0])
        } else {
            c.blocks[address] = newBlock
        }

        return newBlock.data[0], nil
    }

    // 3. Busca da memória principal
    fmt.Printf("Cache %d: Buscando da memória principal\n", c.id)
    value, err := c.bus.memory.Read(address)
    if err != nil {
        return "", fmt.Errorf("erro ao ler da memória: %v", err)
    }

    // Cria novo bloco
    newBlock := &CacheBlock{
        address: address,
        data:    []string{value},
        state:   Exclusive,  // Primeira cache a ter o dado
        time:    c.nextTime,
    }
    c.nextTime++

    // Armazena na cache
    if len(c.blocks) >= c.size {
        c.replaceFIFO(address, value)
    } else {
        c.blocks[address] = newBlock
    }

    return value, nil
}

func (c *Cache) Write(address int, value string) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    fmt.Printf("Cache %d: Escrevendo no endereço %d\n", c.id, address)

    // 1. Invalida cópias em outras caches
    c.invalidateOtherCopies(address)

    // 2. Se o bloco já existe, apenas atualiza
    if block, exists := c.blocks[address]; exists {
        block.data[0] = value
        block.state = Modified
        fmt.Printf("Cache %d: Atualizando bloco existente para Modified\n", c.id)
        return nil
    }

    // 3. Se a cache está cheia, usa FIFO
    if len(c.blocks) >= c.size {
        err := c.replaceFIFO(address, value)
        if err != nil {
            return fmt.Errorf("erro na substituição FIFO: %v", err)
        }
        return nil
    }

    // 4. Se há espaço, adiciona novo bloco
    c.blocks[address] = &CacheBlock{
        address: address,
        data:    []string{value},
        state:   Modified,
        time:    c.nextTime,
    }
    c.nextTime++

    return nil
}

func (c *Cache) replaceFIFO(address int, value string) error {
    var oldestTime int = -1
    var oldestAddress int = -1

    // Encontra bloco mais antigo
    for addr, block := range c.blocks {
        if oldestTime == -1 || block.time < oldestTime {
            oldestTime = block.time
            oldestAddress = addr
        }
    }

    if oldestAddress == -1 {
        return fmt.Errorf("não foi possível encontrar bloco para substituição")
    }

    oldBlock := c.blocks[oldestAddress]
    
    // Write-back se necessário
    if oldBlock.state == Modified {
        err := c.bus.memory.Write(oldBlock.address, oldBlock.data[0])
        if err != nil {
            return fmt.Errorf("erro ao sincronizar com memória: %v", err)
        }
        fmt.Printf("Cache %d: Write-back do bloco %d\n", c.id, oldBlock.address)
    }

    // Remove bloco antigo e adiciona novo
    delete(c.blocks, oldestAddress)
    c.blocks[address] = &CacheBlock{
        address: address,
        data:    []string{value},
        state:   Modified,
        time:    c.nextTime,
    }
    c.nextTime++

    return nil
}

// Métodos de visualização mantidos do código original
func (c *Cache) GetDisplayBlocks() [][]string {
    c.mu.RLock()
    defer c.mu.RUnlock()

    blockSize := c.size / 5
    if blockSize == 0 {
        blockSize = 1
    }

    var blocks [][]string
    for start := 0; start < c.size; start += blockSize {
        end := start + blockSize
        if end > c.size {
            end = c.size
        }

        var block []string
        for i := start; i < end; i++ {
            if cacheBlock, exists := c.blocks[i]; exists {
                block = append(block, fmt.Sprintf(
                    "Linha %d - Endereço: %d, Estado: %s, Dados: %v, Tempo: %d",
                    i, cacheBlock.address, cacheBlock.state.String(), cacheBlock.data, cacheBlock.time,
                ))
            } else {
                block = append(block, fmt.Sprintf("Linha %d - vazia", i))
            }
        }
        blocks = append(blocks, block)
    }

    return blocks
}

func (c *Cache) Display() {
    blocks := c.GetDisplayBlocks()
    for i, block := range blocks {
        fmt.Printf("Cache %d - Página %d:\n", c.id, i+1)
        for _, line := range block {
            fmt.Println(line)
        }
        fmt.Println("Pressione Enter para continuar...")
        fmt.Scanln()
    }
    fmt.Println("Exibição completa!")
}