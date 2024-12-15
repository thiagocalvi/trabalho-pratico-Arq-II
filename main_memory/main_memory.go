package main_memory

import (
	"fmt"
)

type MainMemory struct {
	size int            // Tamanho da memoria principal
	data map[int]string // Dados armazenado na memoria principal, tipo chave, valor. Chave
	// representa o endereço, valor representa o dado guardado no endereço.
}

func NewMemory(size int) *MainMemory {
	memory := &MainMemory{
		size: size,
		data: make(map[int]string, size), // Cria o map com a capacidade inicial de `size`.
	}

	// Preenche o map com valores padrão
	for i := 0; i < size; i++ {
		memory.data[i] = "" // Inicializa cada endereço de memória com uma string vazia.
	}

	return memory
}

// Write escreve um valor na memória em um endereço específico.
func (m *MainMemory) Write(address int, value string) error {
	if address < 0 || address >= m.size {
		return fmt.Errorf("endereço %d está fora dos limites da memória", address)
	}
	m.data[address] = value
	return nil
}

// Read lê um valor da memória em um endereço específico.
func (m *MainMemory) Read(address int) (string, error) {
	if address < 0 || address >= m.size {
		return "", fmt.Errorf("endereço %d está fora dos limites da memória", address)
	}
	return m.data[address], nil
}

// GetDisplayBlocks retorna os blocos da memória a serem exibidos.
func (m *MainMemory) GetDisplayBlocks() [][]string {
	blockSize := m.size / 5
	var blocks [][]string

	for start := 0; start < m.size; start += blockSize {
		end := start + blockSize
		if end > m.size {
			end = m.size
		}

		// Adiciona os dados do bloco atual à lista de blocos.
		var block []string
		for i := start; i < end; i++ {
			block = append(block, fmt.Sprintf("Endereço %d: %s", i, m.data[i]))
		}
		blocks = append(blocks, block)
	}

	return blocks
}

// Display exibe os blocos de memória de forma interativa.
func (m *MainMemory) Display() {
	blocks := m.GetDisplayBlocks()
	for i, block := range blocks {
		fmt.Printf("Bloco %d:\n", i+1)
		for _, line := range block {
			fmt.Println(line)
		}
		fmt.Println("Pressione Enter para continuar...")
		fmt.Scanln() // Espera o Enter
	}
	fmt.Println("Exibição completa!")
}
