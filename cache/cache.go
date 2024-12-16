package cache

// Define um tipo personalizado para o estado da cache com uint8
type CacheBlockState uint8

// Define os estados possíveis para o MESIF (Modified, Exclusive, Shared, Invalid, Forward)
const (
	Modified CacheBlockState = iota
	Exclusive
	Shared
	Invalid
	Forward
)

type CacheBlock struct {
	tag   int             // Tag do bloco
	data  string          // Dado armazenado na linha da cache
	state CacheBlockState // Estado do MESIF (Modified, Exclusive, Shared, Invalid, Forward)
}

type Cache struct {
	size   int                // Tamanho da cache (quantidade de linhas)
	blocks map[int]CacheBlock // Cada bloco representa uma linha da cache, a chave do map é a tag do bloco
	// CacheBlock.tag == chave no map
	queue []int // Fila FIFO para controle de substituição de blocos
}

// Cria uma nova cache com 40% do tamanho da memória principal
func NewCache(mainMemorySize int) *Cache {
	// Calcula o tamanho da cache (40% do tamanho da memória principal)
	cacheSize := int(float64(mainMemorySize) * 0.4)

	// Retorna o endereço da estrutura na memória
	return &Cache{
		size:   cacheSize,                           // Quantidade de "linhas" que a cache tem
		blocks: make(map[int]CacheBlock, cacheSize), // Inicializa o map vazio com tamanho cacheSize
		queue:  make([]int, 0, cacheSize),           // Inicializa a fila FIFO
	}
}

// Write
// Read
