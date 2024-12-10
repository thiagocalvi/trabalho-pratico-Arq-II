package main_memory

import "testing"

// Criação de uma nova MainMemory de tamanho 50
func TestNewMemory(t *testing.T) {
	size := 50
	memory := NewMemory(size)

	if memory == nil {
		t.Fatal("A memória não foi criada corretamente, recebeu nil")
	}

	if memory.size != size {
		t.Errorf("Tamanho da memória incorreto, esperado %d, mas recebeu %d", size, memory.size)
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
