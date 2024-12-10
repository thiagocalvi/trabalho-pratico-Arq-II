package cache

import "testing"

// Criação de uma nova Cache com tamanho 5
func TestNewCache(t *testing.T) {
	size := 5
	cache := NewCache(size)

	if cache == nil {
		t.Fatal("A cache não foi criada corretamente, recebeu nil")
	}

	if cache.size != size {
		t.Errorf("Tamanho da cache incorreto, esperado %d, mas recebeu %d", size, cache.size)
	}

	if len(cache.blocs) != size {
		t.Errorf("Quantidade de blocos incorreta, esperado %d, mas recebeu %d", size, len(cache.blocs))
	}

	for i, block := range cache.blocs {
		if block.tag != 0 {
			t.Errorf("Bloco %d possui tag inicial incorreta, esperado 0, mas recebeu %d", i, block.tag)
		}
		if block.data != "" {
			t.Errorf("Bloco %d possui data inicial incorreta, esperado string vazia, mas recebeu %q", i, block.data)
		}
		if block.state != 0 {
			t.Errorf("Bloco %d possui estado inicial incorreto, esperado 0, mas recebeu %d", i, block.state)
		}
	}
}
