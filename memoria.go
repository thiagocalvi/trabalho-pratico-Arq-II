// memoria.go
package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

// Produto representa um item na memória principal
type Produto struct {
	ID    int     // Identificador único do produto
	Preco float64 // Preço do produto
}

// MemoriaPrincipal é um array de produtos que simula a memória principal do sistema
type MemoriaPrincipal []Produto

// InicializarMemoriaPrincipal preenche a memória principal com produtos aleatórios
// Parâmetros:
//
//	tamanho: número de posições na memória principal
//
// Retorna:
//
//	MemoriaPrincipal inicializada com produtos com valores aleatórios
func InicializarMemoriaPrincipal(tamanho int) MemoriaPrincipal {
	// Cria um gerador de números aleatórios local usando a hora atual como semente
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	memoria := make(MemoriaPrincipal, tamanho)
	for i := 0; i < tamanho; i++ {
		memoria[i] = Produto{
			ID:    i,                 // O ID do produto é o índice na memória
			Preco: r.Float64() * 100, // Preço aleatório entre 0 e 100
		}
	}
	return memoria
}

// ExibirMemoriaPrincipal exibe o conteúdo da memória principal
// Parâmetros:
//
//	memoria: memória principal a ser exibida
func ExibirMemoriaPrincipal(memoria MemoriaPrincipal, conn net.Conn, arquivo *os.File) {
	msg := fmt.Sprintf("=== Memória Principal ===\n")
	conn.Write([]byte(msg))
	arquivo.WriteString(msg)
	for i, produto := range memoria {
		msg = fmt.Sprintf("Linha %d: ID=%d, Preço=%.2f\n", i, produto.ID, produto.Preco)
		conn.Write([]byte(msg))
		arquivo.WriteString(msg)
	}
}
