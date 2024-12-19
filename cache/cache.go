package cache

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
	}

	return cache
}
