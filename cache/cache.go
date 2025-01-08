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
func (c *Cache) ReplaceBlock(address int, value string, memory *main_memory.MainMemory) {
	// Calcula o índice do bloco a ser substituído
	oldestBlockTag := c.queue[0] // O mais antigo é o primeiro na fila (FIFO)
	c.queue = c.queue[1:]        // Remove o mais antigo da fila

	// Verifica o estado do bloco mais antigo
	if oldBlock, exists := c.blocks[oldestBlockTag]; exists && oldBlock.state == Modified {
		// Sincroniza os dados com a memória principal
		err := memory.Write(oldestBlockTag, oldBlock.data[0]) // Escreve o dado na memória
		if err != nil {
			panic(fmt.Sprintf("Erro ao sincronizar bloco com a memória: %v", err))
		}
	}

	// Remove o bloco antigo da cache
	delete(c.blocks, oldestBlockTag)

	// Calcula o novo índice do bloco
	newTag := address % c.size

	// Cria e insere o novo bloco
	newBlock := NewCacheBlock(newTag, len(value))
	newBlock.data[0] = value  // Adiciona o dado ao novo bloco
	newBlock.state = Modified // Define o estado como `Modified`

	c.blocks[newTag] = newBlock       // Adiciona o novo bloco à cache
	c.queue = append(c.queue, newTag) // Insere o novo bloco na fila
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

func (c *Cache) Read(address int, memory *main_memory.MainMemory) string {
	tag := address % c.size
	fmt.Printf("Lendo endereço %d (tag: %d)...\n", address, tag)

	// Verifica se o bloco existe na cache
	if block, exists := c.blocks[tag]; exists {
		if len(block.data) > 0 {
			fmt.Printf("Cache hit: Dados encontrados no bloco %d.\n", tag)
			return block.data[0]
		} else {
			fmt.Printf("Erro: Bloco encontrado, mas 'data' está vazio.\n")
		}
	}

	// Cache miss: Lê da memória principal
	fmt.Printf("Cache miss: Lendo da memória principal.\n")
	memValue, err := memory.Read(address)
	if err != nil {
		panic(fmt.Sprintf("Erro ao ler da memória principal: %v", err))
	}

	// Substitui ou adiciona o bloco na cache
	c.ReplaceBlock(address, memValue, memory)
	fmt.Printf("Bloco substituído ou adicionado na cache (tag: %d).\n", tag)

	return memValue
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
