package cache

type CacheBlock struct {
	tag   int    // Tag do bloco
	data  string // Dado armazenado na linha da cache
	state int    // Estado do MESIF
}

type Cache struct {
	size  int // Tamanho da cache (quantidade de linhas)
	blocs []CacheBlock
}

// Cria uma nova cache
func NewCache(size int) *Cache {
	return &Cache{ // Retorna o endereço da estrutura na memória
		size:  size,
		blocs: make([]CacheBlock, size),
	}
}
