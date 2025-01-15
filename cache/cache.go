package cache

import (
    "fmt"
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

type CacheBlock struct {
    address int             // Endereço do bloco na memória principal
    data    []string        // Dados armazenados na linha da cache
    state   CacheBlockState // Estado do MESIF
    time    int            // Para implementação do FIFO
}

type Cache struct {
    size      int                 // Número de linhas na cache
    lineSize  int                 // Número de dados por linha
    blocks    map[int]*CacheBlock // Mapeamento de linhas para blocos
    nextTime  int                 // Contador para FIFO
}

func NewCache(mainMemorySize int) *Cache {
    cacheSize := int(float64(mainMemorySize) * 0.4) // 40% do tamanho da memória principal
    lineSize := int(float64(cacheSize) * 0.2)       // 20% do tamanho da cache

    return &Cache{
        size:     cacheSize,
        lineSize: lineSize,
        blocks:   make(map[int]*CacheBlock),
        nextTime: 0,
    }
}

func (c *Cache) read(address int, memory *main_memory.MainMemory) string {
    fmt.Printf("Lendo endereço %d...\n", address)

    // Procura o endereço em qualquer linha da cache (mapeamento associativo)
    for idx, block := range c.blocks {
        if block.address == address {
            fmt.Printf("Cache hit: Dados encontrados na linha %d\n", idx)
            if block.state == Invalid {
                // Se o bloco estiver inválido, busca da memória principal
                memValue, err := memory.Read(address)
                if err != nil {
                    panic(fmt.Sprintf("Erro ao ler da memória principal: %v", err))
                }
                block.data[0] = memValue
                block.state = Exclusive // Primeiro acesso após inválido
            }
            return block.data[0]
        }
    }

    // Cache miss: Lê da memória principal
    fmt.Printf("Cache miss: Lendo da memória principal.\n")
    memValue, err := memory.Read(address)
    if err != nil {
        panic(fmt.Sprintf("Erro ao ler da memória principal: %v", err))
    }

    // Se houver espaço na cache
    if len(c.blocks) < c.size {
        // Encontra primeira linha disponível
        for i := 0; i < c.size; i++ {
            if _, exists := c.blocks[i]; !exists {
                c.blocks[i] = &CacheBlock{
                    address: address,
                    data:    []string{memValue},
                    state:   Exclusive,
                    time:    c.nextTime,
                }
                c.nextTime++
                return memValue
            }
        }
    }

    // Caso não haja espaço, usa FIFO para substituição
    c.replaceFIFO(address, memValue, memory)
    return memValue
}

func (c *Cache) replaceFIFO(address int, value string, memory *main_memory.MainMemory) {
    var oldestTime int = -1
    var oldestIdx int = -1

    // Encontra o bloco mais antigo
    for idx, block := range c.blocks {
        if oldestTime == -1 || block.time < oldestTime {
            oldestTime = block.time
            oldestIdx = idx
        }
    }

    oldBlock := c.blocks[oldestIdx]
    
    // Se o bloco mais antigo estiver modificado, sincroniza com a memória
    if oldBlock.state == Modified {
        err := memory.Write(oldBlock.address, oldBlock.data[0])
        if err != nil {
            panic(fmt.Sprintf("Erro ao sincronizar com a memória principal: %v", err))
        }
    }

    // Substitui o bloco
    c.blocks[oldestIdx] = &CacheBlock{
        address: address,
        data:    []string{value},
        state:   Exclusive,
        time:    c.nextTime,
    }
    c.nextTime++
}

func (c *Cache) Write(address int, value string, memory *main_memory.MainMemory) error {
    fmt.Printf("Escrevendo no endereço %d...\n", address)

    // Procura o bloco em toda a cache (mapeamento associativo)
    for _, block := range c.blocks {
        if block.address == address {
            block.data[0] = value
            block.state = Modified
            return nil
        }
    }

    // Se não encontrou o bloco na cache
    if len(c.blocks) < c.size {
        // Procura primeira linha disponível
        for i := 0; i < c.size; i++ {
            if _, exists := c.blocks[i]; !exists {
                c.blocks[i] = &CacheBlock{
                    address: address,
                    data:    []string{value},
                    state:   Modified,
                    time:    c.nextTime,
                }
                c.nextTime++
                return nil
            }
        }
    }

    // Se não há espaço, usa FIFO
    c.replaceFIFO(address, value, memory)
    return nil
}

func (c *Cache) GetDisplayBlocks() [][]string {
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
        fmt.Printf("Página %d:\n", i+1)
        for _, line := range block {
            fmt.Println(line)
        }
        fmt.Println("Pressione Enter para continuar...")
        fmt.Scanln()
    }
    fmt.Println("Exibição completa!")
}