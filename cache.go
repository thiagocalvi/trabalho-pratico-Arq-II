package main

import (
	"fmt"
	"net"
	"os"
)

// Estado representa os possíveis estados de uma linha de cache no protocolo MESIF
type Estado string

const (
	Modify    Estado = "M" // Dado modificado e não sincronizado com a memória
	Exclusive Estado = "E" // Dado exclusivo (única cópia válida)
	Shared    Estado = "S" // Dado compartilhado entre múltiplas caches
	Invalid   Estado = "I" // Linha inválida/vazia
	Forward   Estado = "F" // Cache designada como fonte oficial para o dado
)

// LinhaCache representa uma entrada na memória cache
type LinhaCache struct {
	Tag    int     // Endereço do bloco na memória principal
	Dado   Produto // Dado armazenado
	Estado Estado  // Estado atual da linha
}

// Cache representa a memória cache de um processador
type Cache []LinhaCache

// cacheRegistry mantém um registro global de quais caches possuem cada endereço
var cacheRegistry = make(map[int][]*Cache)

// InicializarCache cria uma nova cache com o tamanho especificado
// Parâmetros:
//
//	tamanho: número de linhas na cache
//
// Retorna:
//
//	Cache inicializada com todas as linhas no estado Invalid
func InicializarCache(tamanho int) Cache {
	cache := make(Cache, tamanho)
	for i := range cache {
		cache[i] = LinhaCache{Tag: -1, Estado: Invalid}
	}
	return cache
}

// ExibirCache mostra o conteúdo atual de uma cache
// Parâmetros:
//
//	cache: cache a ser exibida
//	processador: número do processador para identificação
//	conn: conexão com o servidor
//	arquivo: aquivo de log
func ExibirCache(cache Cache, processador int, conn net.Conn, arquivo *os.File) {
	msg := fmt.Sprintf("\n=== Cache do processador P%d ===\n", processador)
	conn.Write([]byte(msg))
	arquivo.WriteString(msg)
	for i, linha := range cache {
		msg = fmt.Sprintf("Linha %d: Tag=%d, ID=%d, Preço=%.2f, Estado=%s\n", i,
			linha.Tag, linha.Dado.ID, linha.Dado.Preco, linha.Estado)
		conn.Write([]byte(msg))
		arquivo.WriteString(msg)
	}
}

// EncontrarLinhaParaSubstituir implementa a política de substituição FIFO com write-back
// Parâmetros:
//
//	cache: cache onde será feita a substituição
//	memoria: referência à memória principal para write-back
//	conn: conexão com o servidor
//	arquivo: aquivo de log
//
// Retorna:
//
//	Índice da linha a ser substituída
func EncontrarLinhaParaSubstituir(cache Cache, memoria MemoriaPrincipal, conn net.Conn, arquivo *os.File) int {
	// Prioriza linhas inválidas
	for i, linha := range cache {
		if linha.Estado == Invalid {
			return i
		}
	}

	// Política FIFO: substitui a primeira linha (posição 0)
	linhaSubstituida := 0
	if cache[0].Estado == Modify {
		// Write-back: escreve na memória se estiver modificado
		memoria[cache[0].Tag] = cache[0].Dado
		msg := fmt.Sprintf("Write-Back: Dado %d escrito na memória\n", cache[0].Tag)
		conn.Write([]byte(msg))
		arquivo.WriteString(msg)
	}
	return linhaSubstituida
}

// InvalidarOutrasCaches invalida todas as cópias do endereço em outras caches
// Parâmetros:
//
//	endereco: endereço de memória a ser invalidado
//	cacheAtual: cache que está realizando a operação (não será invalidada)
//	conn: conexão com o servidor
//	arquivo: aquivo de log
func InvalidarOutrasCaches(endereco int, cacheAtual *Cache, conn net.Conn, arquivo *os.File, memoria MemoriaPrincipal) {
	if caches, ok := cacheRegistry[endereco]; ok {
		for _, c := range caches {
			if c != cacheAtual {
				for i := range *c {
					if (*c)[i].Tag == endereco {
						// Se o estado for Modify atualiza na memória principal
						if (*c)[i].Estado == Modify {
							memoria[endereco] = (*c)[endereco-1].Dado
							msg := fmt.Sprintf("Dado escrito na memória principal endereço %d\n", endereco)
							conn.Write([]byte(msg))
							arquivo.WriteString(msg)
						}
						// Invalida a linha da cache
						(*c)[i].Estado = Invalid
						msg := fmt.Sprintf("Cache invalidada: Endereço %d\n", endereco)
						conn.Write([]byte(msg))
						arquivo.WriteString(msg)

					}
				}
			}
		}
	}
}

// AtualizarEstadoForward gerencia a designação do estado Forward conforme protocolo MESIF
// Parâmetros:
//
//	endereco: endereço de memória sendo acessado
//	cache: cache que está sendo atualizada
//	conn: conexão com o servidor
//	arquivo: aquivo de log
func AtualizarEstadoForward(endereco int, cache *Cache, conn net.Conn, arquivo *os.File) {
	if len(cacheRegistry[endereco]) == 1 {
		// Caso único: marca como Exclusive
		for i := range *cache {
			if (*cache)[i].Tag == endereco {
				(*cache)[i].Estado = Exclusive
			}
		}
	} else {
		// Verifica se já existe uma Forward
		designada := false
		for _, c := range cacheRegistry[endereco] {
			for i := range *c {
				if (*c)[i].Tag == endereco && (*c)[i].Estado == Forward {
					designada = true
					break
				}
			}
		}

		// Designa nova Forward se necessário
		if !designada {
			for i := range *cache {
				if (*cache)[i].Tag == endereco {
					(*cache)[i].Estado = Forward
					msg := fmt.Sprintf("Cache designada como Forward para endereço %d\n", endereco)
					conn.Write([]byte(msg))
					arquivo.WriteString(msg)
				}
			}
		}
	}
}

// LerDado simula uma operação de leitura na cache seguindo o protocolo MESIF
// Parâmetros:
//
//	processador: identificador do processador
//	endereco: endereço de memória a ser lido
//	memoria: referência à memória principal
//	cache: cache do processador
//	conn: conexão com o servidor
//	arquivo: aquivo de log
func LerDado(processador int, endereco int, memoria MemoriaPrincipal, cache *Cache, conn net.Conn, arquivo *os.File) {
	// Read Hit: dado encontrado localmente
	for _, linha := range *cache {
		if linha.Tag == endereco && linha.Estado != Invalid {
			msg := fmt.Sprintf("P%d: Read Hit (RH) no endereço %d\n", processador, endereco)
			conn.Write([]byte(msg))
			arquivo.WriteString(msg)
			return
		}
	}

	// Read Miss: dado não encontrado localmente
	msg := fmt.Sprintf("P%d: Read Miss (RM) no endereço %d\n", processador, endereco)
	conn.Write([]byte(msg))
	arquivo.WriteString(msg)

	// Busca em outras caches
	var fonte *Cache
	for _, c := range cacheRegistry[endereco] {
		for _, linha := range *c {
			if linha.Tag == endereco && linha.Estado != Invalid {
				fonte = c
				break
			}
		}
	}

	var dado Produto
	if fonte != nil {
		// Carrega de outra cache
		msg := fmt.Sprintf("P%d: Dado carregado de outra cache\n", processador)
		conn.Write([]byte(msg))
		arquivo.WriteString(msg)
		for _, linha := range *fonte {
			if linha.Tag == endereco {
				dado = linha.Dado
				(*fonte)[endereco-1].Estado = Shared
				break
			}
		}
	} else {
		// Carrega da memória principal
		dado = memoria[endereco]
		msg := fmt.Sprintf("P%d: Dado carregado da memória pricipal\n", processador)
		conn.Write([]byte(msg))
		arquivo.WriteString(msg)

		msg = fmt.Sprintf("P%d: Dado carregado: ID=%d, Preço=%.2f \n", processador, dado.ID, dado.Preco)
		conn.Write([]byte(msg))
		arquivo.WriteString(msg)
	}

	// Substituição de linha
	idx := EncontrarLinhaParaSubstituir(*cache, memoria, conn, arquivo)
	(*cache)[idx] = LinhaCache{
		Tag:    endereco,
		Dado:   dado,
		Estado: Shared,
	}

	// Atualiza registros e estados
	cacheRegistry[endereco] = append(cacheRegistry[endereco], cache)
	AtualizarEstadoForward(endereco, cache, conn, arquivo)

}

// EscreverDado simula uma operação de escrita na cache seguindo o protocolo MESIF
// Parâmetros:
//
//	processador: identificador do processador
//	endereco: endereço de memória a ser escrito
//	valor: novo valor a ser armazenado
//	memoria: referência à memória principal
//	cache: cache do processador
//	conn: conexão com o servidor
//	arquivo: aquivo de log
func EscreverDado(processador int, endereco int, valor float64, memoria MemoriaPrincipal, cache *Cache, conn net.Conn, arquivo *os.File) {
	// Write Hit: dado encontrado localmente
	for i, linha := range *cache {
		if linha.Tag == endereco && linha.Estado != Invalid {
			msg := fmt.Sprintf("P%d: Write Hit (WH) no endereço %d\n", processador, endereco)
			conn.Write([]byte(msg))
			arquivo.WriteString(msg)
			(*cache)[i].Dado.Preco = valor
			(*cache)[i].Estado = Modify
			InvalidarOutrasCaches(endereco, cache, conn, arquivo, memoria)
			return
		}
	}

	// Write Miss: dado não encontrado localmente
	msg := fmt.Sprintf("P%d: Write Miss (WM) no endereço %d\n", processador, endereco)
	conn.Write([]byte(msg))
	arquivo.WriteString(msg)

	// Carrega o dado da memória e atualiza o valor
	dado := memoria[endereco]
	dado.Preco = valor

	// Substituição de linha
	idx := EncontrarLinhaParaSubstituir(*cache, memoria, conn, arquivo)
	(*cache)[idx] = LinhaCache{
		Tag:    endereco,
		Dado:   dado,
		Estado: Modify,
	}

	// Atualiza registros e estados
	cacheRegistry[endereco] = append(cacheRegistry[endereco], cache)
	InvalidarOutrasCaches(endereco, cache, conn, arquivo, memoria)
}
