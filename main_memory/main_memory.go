package main_memory

type MainMemory struct {
	size int            // Tamanho da memoria principal
	data map[int]string // Dados armazenado na memoria principal, tipo chave, valor. Chave
	// representa o endereço, valor representa o dado guardado no endereço.
}

func NewMemory(size int) *MainMemory {
	memory := &MainMemory{
		size: size,
		data: make(map[int]string, size), // Cria o mapa com a capacidade inicial de `size`.
	}

	// Preenche o mapa com valores padrão
	for i := 0; i < size; i++ {
		memory.data[i] = "" // Inicializa cada endereço de memória com uma string vazia.
	}

	return memory
}
