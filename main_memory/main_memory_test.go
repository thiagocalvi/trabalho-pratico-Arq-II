package main_memory

import (
	"fmt"
	"testing"
)

// Criação de uma nova MainMemory de tamanho 50
func TestNewMemory(t *testing.T) {
	size := 50
	memory := NewMemory(size)

	if memory == nil {
		t.Fatal("A memória não foi criada corretamente, recebeu nil")
	}

	if memory.Size != size {
		t.Errorf("Tamanho da memória incorreto, esperado %d, mas recebeu %d", size, memory.Size)
	}

	if len(memory.data) != size {
		t.Errorf("O mapa de dados não contém o número esperado de elementos, esperado %d, mas recebeu %d", size, len(memory.data))
	}

	// Verifica se todas as posições estão inicializadas com string vazia
	for i := 0; i < size; i++ {
		if value, exists := memory.data[i]; !exists || value != "" {
			t.Errorf("Endereço %d não foi inicializado corretamente, esperado string vazia, mas recebeu %q (exists: %v)", i, value, exists)
		}
	}
}

func TestWrite(t *testing.T) {
	size := 50
	memory := NewMemory(size)

	// Testa uma escrita válida.
	address := 10
	value := "teste"
	err := memory.Write(address, value)
	if err != nil {
		t.Errorf("Erro ao escrever na memória: %v", err)
	}

	// Verifica se o valor foi escrito corretamente.
	if memory.data[address] != value {
		t.Errorf("Valor incorreto na posição %d, esperado %q, mas recebeu %q", address, value, memory.data[address])
	}

	// Testa escrita fora dos limites (endereço negativo).
	err = memory.Write(-1, "fora dos limites")
	if err == nil {
		t.Error("Escrita fora dos limites (endereço negativo) não retornou erro")
	}

	// Testa escrita fora dos limites (endereço maior que o tamanho).
	err = memory.Write(size, "fora dos limites")
	if err == nil {
		t.Error("Escrita fora dos limites (endereço maior que o tamanho) não retornou erro")
	}
}

func TestRead(t *testing.T) {
	size := 50
	memory := NewMemory(size)

	// Configura um valor para teste.
	address := 20
	expectedValue := "valor de teste"
	memory.data[address] = expectedValue

	// Testa leitura válida.
	value, err := memory.Read(address)
	if err != nil {
		t.Errorf("Erro ao ler da memória: %v", err)
	}

	if value != expectedValue {
		t.Errorf("Valor incorreto lido da posição %d, esperado %q, mas recebeu %q", address, expectedValue, value)
	}

	// Testa leitura fora dos limites (endereço negativo).
	_, err = memory.Read(-1)
	if err == nil {
		t.Error("Leitura fora dos limites (endereço negativo) não retornou erro")
	}

	// Testa leitura fora dos limites (endereço maior que o tamanho).
	_, err = memory.Read(size)
	if err == nil {
		t.Error("Leitura fora dos limites (endereço maior que o tamanho) não retornou erro")
	}
}

func TestGetDisplayBlocks(t *testing.T) {
	size := 10
	memory := NewMemory(size)

	// Preenche alguns valores na memória.
	for i := 0; i < size; i++ {
		memory.data[i] = fmt.Sprintf("valor %d", i)
	}

	// Gera os blocos de exibição.
	blocks := memory.GetDisplayBlocks()

	// Valida que os blocos foram gerados corretamente.
	expectedBlockCount := 5 // 20% da memória por bloco.
	if len(blocks) != expectedBlockCount {
		t.Errorf("Número incorreto de blocos gerados, esperado %d, mas recebeu %d", expectedBlockCount, len(blocks))
	}

	// Verifica os valores de cada bloco.
	expectedBlockSize := size / expectedBlockCount
	for i, block := range blocks {
		if len(block) != expectedBlockSize {
			t.Errorf("Tamanho incorreto do bloco %d, esperado %d, mas recebeu %d", i+1, expectedBlockSize, len(block))
		}

		// Verifica o conteúdo do bloco.
		start := i * expectedBlockSize
		for j, line := range block {
			expectedLine := fmt.Sprintf("Endereço %d: valor %d", start+j, start+j)
			if line != expectedLine {
				t.Errorf("Conteúdo incorreto no bloco %d, linha %d, esperado %q, mas recebeu %q", i+1, j+1, expectedLine, line)
			}
		}
	}
}
