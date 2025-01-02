package cache

import (
	"fmt"
	"trabalho_pratico/main_memory"
)

// Define os estados possíveis para o MESIF (Modified, Exclusive, Shared, Invalid, Forward)
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
	tag   int             // Tag do bloco
	data  []string        // Dados armazenados na linha da cache
	state CacheBlockState // Estado do MESIF
}

type Cache struct {
	size     int                 // Número de linhas na cache
	lineSize int                 // Número de dados por linha
	blocks   map[int]*CacheBlock // Mapeamento de tags para blocos
	queue    []int               // Fila FIFO para substituição
}

// Cria um novo bloco de cache
func NewCacheBlock(cacheSize, tag int) *CacheBlock {
    lineSize := int(float64(cacheSize) * 0.2) // 20% do tamanho da cache
    return &CacheBlock{
        tag:   tag,
        data:  make([]string, lineSize), // Configurando capacidade correta
        state: Invalid,
    }
}

func NewCache(mainMemorySize int) *Cache {
    cacheSize := int(float64(mainMemorySize) * 0.4) // 40% do tamanho da memória principal

    cache := &Cache{
        size:   cacheSize,
        blocks: make(map[int]*CacheBlock, cacheSize),
        queue:  make([]int, 0, cacheSize), // Inicializa a fila vazia com capacidade máxima
    }

    // Preenche o mapa de blocos
    for i := 0; i < cacheSize; i++ {
        cache.blocks[i] = &CacheBlock{
            tag:   i,
            data:  make([]string, 0), // Dados inicialmente vazios
            state: Invalid,           // Estado inicial: Inválido
        }
    }

    return cache
}

// Substitui o bloco mais antigo (FIFO) e atualiza o conteúdo
func (c *Cache) ReplaceBlock(tag int, newData string, memory *main_memory.MainMemory) {
    // Verifica se a fila está cheia
    if len(c.queue) >= c.size {
        // Remove o bloco mais antigo
        oldestTag := c.queue[0]
        c.queue = c.queue[1:]

        if oldestBlock, exists := c.blocks[oldestTag]; exists {
            if oldestBlock.state == Modified {
                // Sincroniza com a memória principal
                memory.Write(oldestTag, oldestBlock.data[0])
            }
            delete(c.blocks, oldestTag) // Remove o bloco da cache
        }
    }

    // Adiciona o novo bloco à cache
    newBlock := &CacheBlock{
        tag:   tag,
        data:  []string{newData},
        state: Shared, // Estado inicial como Shared
    }
    c.blocks[tag] = newBlock
    c.queue = append(c.queue, tag) // Atualiza a fila
}

func (c *Cache) Write(address int, value string, memory *main_memory.MainMemory) error {
    tag := address % c.size // Calcula a tag com base no tamanho da cache
    block, exists := c.blocks[tag]

    // Se o bloco não existir ou estiver inválido, cria ou atualiza
    if !exists || block.state == Invalid {
        if exists && block.state == Modified {
            // Sincroniza o bloco modificado com a memória principal antes de sobrescrever
            if err := memory.Write(block.tag, block.data[0]); err != nil {
                return fmt.Errorf("erro ao sincronizar bloco modificado: %w", err)
            }
        }

        // Cria um novo bloco para a tag atual
        block = &CacheBlock{
            tag:   tag,
            data:  make([]string, 1), // Aloca espaço para o dado
            state: Modified,
        }
        c.blocks[tag] = block
    }

    // Atualiza o dado no bloco
    if len(block.data) == 0 {
        block.data = append(block.data, value) // Inicializa se necessário
    } else {
        block.data[0] = value
    }

    block.state = Modified // Atualiza o estado para modificado

    // Adiciona o bloco na fila para gerenciamento
    c.queue = append(c.queue, tag)

    return nil
}

func (c *Cache) Read(address int, memory *main_memory.MainMemory) (string, error) {
    tag := address % c.size // Calcula a tag baseada no tamanho da cache
    block, exists := c.blocks[tag]

    if exists && block.state != Invalid {
        // Cache hit
        if len(block.data) > 0 {
            return block.data[0], nil // Retorna o dado diretamente
        }
    }

    // Cache miss: busca o dado na memória principal
    value, err := memory.Read(address)
    if err != nil {
        return "", fmt.Errorf("erro ao ler da memória principal: %w", err)
    }

    // Atualiza ou cria o bloco na cache
    if !exists || block == nil {
        block = &CacheBlock{
            tag:   tag,
            data:  make([]string, 1), // Inicializa a lista de dados
            state: Shared,
        }
        c.blocks[tag] = block // Adiciona o bloco à cache
    }

    block.data[0] = value    // Atualiza o dado no bloco
    block.state = Shared     // Define o estado como Shared

    // Atualiza a fila para gerenciamento da cache
    c.queue = append(c.queue, tag)
    if len(c.queue) > c.size {
        oldestTag := c.queue[0]
        c.queue = c.queue[1:] // Remove o bloco mais antigo da fila

        if oldestBlock, exists := c.blocks[oldestTag]; exists && oldestBlock.state == Modified {
            // Sincroniza o bloco mais antigo com a memória principal se necessário
            memory.Write(oldestTag, oldestBlock.data[0])
        }
        delete(c.blocks, oldestTag) // Remove o bloco da cache
    }

    return value, nil
}

func (c *Cache) GetDisplayBlocks() [][]string {
	blockSize := c.size / 5
	if blockSize == 0 {
		blockSize = 1 // Garante que haverá pelo menos um bloco por página
	}

	var blocks [][]string
	for start := 0; start < c.size; start += blockSize {
		end := start + blockSize
		if end > c.size {
			end = c.size
		}

		// Adiciona os dados do bloco atual à lista de blocos
		var block []string
		for i := start; i < end; i++ {
			if cacheBlock, exists := c.blocks[i]; exists {
				block = append(block, fmt.Sprintf(
					"Bloco %d - Tag: %d, Estado: %s, Dados: %v",
					i, cacheBlock.tag, cacheBlock.state.String(), cacheBlock.data,
				))
			} else {
				block = append(block, fmt.Sprintf("Bloco %d - vazio", i))
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
		fmt.Scanln() // Espera o Enter
	}
	fmt.Println("Exibição completa!")
}

// Modificar a lógica do block.data[0] caso seja necessário.