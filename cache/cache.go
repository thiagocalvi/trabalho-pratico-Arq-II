package cache

import "trabalho_pratico/main_memory"

// Define um tipo personalizado para o estado de uma linha da cache com uint8
type CacheBlockState uint8

// Define os estados possíveis para o MESIF (Modified, Exclusive, Shared, Invalid, Forward)
const (
	Modified CacheBlockState = iota
	Exclusive
	Shared
	Invalid
	Forward
)

// "Linha" da cache
type CacheBlock struct {
	tag   int             // Tag do bloco
	data  []string        // Dados armazenado na linha da cache, cada posição do vetor corresponde a um dado
	state CacheBlockState // Estado do MESIF (Modified, Exclusive, Shared, Invalid, Forward)
}

type Cache struct {
	size     int                // Tamanho da cache (quantidade de linhas)
	lineSize int                // Quantidade de dados armazenado por linha
	blocks   map[int]CacheBlock // Blocos da cache
	queue    []int              // Fila FIFO para controle de substituição de blocos
}

// Cria um novo bloco da cache
func NewCacheBlock(cacheSize int, tagNumber int) *CacheBlock {
	// Quantidade de dados que uma linha da cache armazena é de 20% do tamanho total da cache
	// a valiar se esse é o melhor valor
	cacheLineSize := int(float64(cacheSize) * 0.2)

	// Retorna o endereço da estrutura na memória
	return &CacheBlock{
		tag:   tagNumber,
		data:  make([]string, cacheLineSize),
		state: Invalid, // O estado inicial da linha é definido como Invalido
	}
}

// Cria uma nova cache com 40% do tamanho da memória principal
func NewCache(mainMemorySize int) *Cache {
	// Calcula o tamanho da cache (40% do tamanho da memória principal)
	cacheSize := int(float64(mainMemorySize) * 0.4)

	// Retorna o endereço da estrutura na memória
	cache := &Cache{
		size:   cacheSize,                           // Quantidade de "linhas" que a cache tem
		blocks: make(map[int]CacheBlock, cacheSize), // Inicializa o map vazio com tamanho cacheSize
		queue:  make([]int, 0, cacheSize),           // Inicializa a fila FIFO
	}

	// Adiciona os blocos da cache na *cache*
	for x := 0; x < cacheSize; x++ {
		block := NewCacheBlock(cacheSize, x) // *x* é a tag do bloco
		cache.blocks[x] = *block
	//	cache.queue = append(cache.queue, x) // Adiciona à fila FIFO
	}

	return cache
}

// Escreve um *value* na cache
//func (c *Cache) Write(address int, value string, memory *main_memory.MainMemory) error {
//	return nil
//}

// Lê um valor da cache
//func (c *Cache) Read(address int) error {
//	return nil
//}

// Função replaceBlock(FIFO)
func (c *Cache) replaceBlock(address int, value string, memory *main_memory.MainMemory) {
	// Remove o bloco mais antigo da fila FIFO
	oldest := c.queue[0]
	c.queue = c.queue[1:]

	// Sincroniza o bloco modificado com a memória principal (write-back)
	if c.blocks[oldest].state == Modified {
		for i, val := range c.blocks[oldest].data {
			memory.Write(oldest+i, val) // Grava cada dado no endereço correspondente
		}
	}

	// Reutilização do bloco removido(não sei se precisa) com atualização
	// Atualiza o bloco com os novos dados
	c.blocks[oldest] = CacheBlock{
		tag:   address,
		data:  []string{value}, // Atualiza o valor na linha
		state: Modified,
	}

	// Adiciona o bloco à fila FIFO
	c.queue = append(c.queue, oldest)
}

// Função Write(completar a que estava acima)
func (c *Cache) Write(address int, value string, memory *main_memory.MainMemory) error {
	// Calcula a tag e o índice do bloco na cache
	tag := address % c.size

	// Verifica se o bloco já está na cache
	if block, exists := c.blocks[tag]; exists && block.state != Invalid {
		// Write Hit: Atualiza o bloco existente
		block.data[0] = value
		block.state = Modified
		c.blocks[tag] = block
		return nil
	}

	// Write Miss: Caso o bloco não esteja na cache, realiza a substituição
	c.replaceBlock(address, value, memory)
	return nil
}

// Função Read(completar a que estava acima)
func (c *Cache) Read(address int, memory *main_memory.MainMemory) (string, error) {
	// Calcula a tag do bloco
	tag := address % c.size

	// Verifica se o bloco está na cache
	if block, exists := c.blocks[tag]; exists && block.state != Invalid {
		// Read Hit: Retorna o dado diretamente
		return block.data[0], nil
	}

	// Read Miss: Carrega o dado da memória principal
	value := memory.Read(address)

	// Adiciona o bloco na cache
	c.replaceBlock(address, value, memory)

	return value, nil
}

// Modificar a lógica do block.data[0] caso seja necessário.