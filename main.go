// main.go
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func main() {
	// Se o programa foi chamado com o argumento "server", ele executa o servidor
	if len(os.Args) > 1 && os.Args[1] == "server" {
		runServer()
		return
	}

	// Se não, executa o cliente normalmente
	go func() {
		clearScreen()
		err := startNewTerminal("server") // Inicia o servidor em um novo terminal
		if err != nil {
			fmt.Println("Erro ao iniciar o servidor:", err)
		}
	}()

	// Aguardar um tempo para garantir que o servidor foi iniciado
	time.Sleep(2 * time.Second)

	// Executa o cliente no terminal principal
	runClient()
}

// Função que roda o servidor
func runServer() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		return
	}
	defer ln.Close()

	fmt.Println("[Servidor] Rodando na porta 8080. Aguardando conexões...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("[Servidor] Erro ao aceitar conexão:", err)
			continue
		}

		// Lidar com o cliente em uma goroutine
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	fmt.Println("[Servidor] Conexão estabelecida com um cliente.")

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("[Servidor] Conexão encerrada:", err)
			break
		}
		fmt.Printf(message)
		conn.Write([]byte("Mensagem recebida pelo servidor\n"))
	}
}

// Abre um novo terminal para executar o servidor
func startNewTerminal(mode string) error {
	fmt.Println("========== Conectado ao servidor. ==========")
	var cmd *exec.Cmd

	// Identifica o OS atual e define o cmd
	switch runtime.GOOS {
	// Linux
	case "linux":
		cmd = exec.Command("xterm", "-e", os.Args[0], mode)
	// Windows
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", os.Args[0], mode) //Testar
	// MacOs
	case "darwin":
		// Acho que vai ser meio dificil de testar esse
		cmd = exec.Command("osascript", "-e", fmt.Sprintf("tell application \"Terminal\" to do script \"%s %s\"", os.Args[0], mode))
	}

	// Iniciar o segundo cmd
	return cmd.Start()
}

// Função para limpar o terminal
func clearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// Função para conectar ao servidor
func connectToServer() (net.Conn, error) {
	for i := 0; i < 5; i++ {
		conn, err := net.Dial("tcp", "localhost:8080")
		if err == nil {
			return conn, nil
		}
		fmt.Println("Tentativa de reconexão...")
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("não foi possível conectar ao servidor após várias tentativas")
}

// Função que roda o cliente
func runClient() {

	conn, err := connectToServer()
	if err != nil {
		fmt.Println("-> Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	// Abrir o arquivo de log em mode de escrita
	// O arquivo de log é criado no diretório onde está o executavel
	file, err := os.OpenFile("trab-arq-logs.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo:", err)
		return
	}
	defer file.Close()

	// Inicializar a memória principal com 50 produtos
	memoria := InicializarMemoriaPrincipal(50)

	// Inicializar as caches dos três processadores, cada um com 5 linhas
	cacheP1 := InicializarCache(5)

	cacheP2 := InicializarCache(5)

	cacheP3 := InicializarCache(5)

	interacao := 1
	// Loop principal de interação
	for {
		file.WriteString(fmt.Sprintf("------------------------ Interação %d ------------------------\n", interacao))
		interacao++

		clearScreen()
		// Exibir o menu de opções
		fmt.Println("\nEscolha uma opção:")
		fmt.Println("1. Ler dado")
		fmt.Println("2. Escrever dado")
		fmt.Println("3. Exibir memória principal")
		fmt.Println("4. Exibir caches")
		fmt.Println("5. Sair (Ctrl + c)")

		var opcao int
		fmt.Print("Opção: ")
		fmt.Scan(&opcao)

		conn.Write([]byte("-------------------------------------------------\n"))
		switch opcao {
		case 1:
			// Operação de leitura
			msg := "Operação de leitura de dado selecionada\n"
			conn.Write([]byte(msg))
			file.WriteString(msg)

			var processador, endereco int
			fmt.Print("Escolha o processador (1, 2 ou 3): ")
			fmt.Scan(&processador)

			msg = fmt.Sprintf("Processador P%d selecionado\n", processador)
			conn.Write([]byte(msg))
			file.WriteString(msg)

			fmt.Print("Digite o endereço de memória (0-49): ")
			fmt.Scan(&endereco)

			if endereco < 0 || endereco >= len(memoria) {
				fmt.Println("Endereço inválido!")
				continue
			}

			switch processador {
			case 1:
				LerDado(1, endereco, memoria, &cacheP1, conn, file)
			case 2:
				LerDado(2, endereco, memoria, &cacheP2, conn, file)
			case 3:
				LerDado(3, endereco, memoria, &cacheP3, conn, file)
			default:
				fmt.Println("Processador inválido!")
				continue
			}

		case 2:
			// Operação de escrita
			msg := "Operação de escrita de dado selecionada\n"
			conn.Write([]byte(msg))
			file.WriteString(msg)

			var processador, endereco int
			var valor float64
			fmt.Print("Escolha o processador (1, 2 ou 3): ")
			fmt.Scan(&processador)

			msg = fmt.Sprintf("Processador P%d selecionado\n", processador)
			conn.Write([]byte(msg))
			file.WriteString(msg)

			fmt.Print("Digite o endereço de memória (0-49): ")
			fmt.Scan(&endereco)

			msg = fmt.Sprintf("Endereço de memória %d selecionado\n", endereco)
			conn.Write([]byte(msg))
			file.WriteString(msg)

			fmt.Print("Digite o novo valor: ")
			fmt.Scan(&valor)

			msg = fmt.Sprintf("Valor %f sendo escrito no endereço %d\n", valor, endereco)
			conn.Write([]byte(msg))
			file.WriteString(msg)

			if endereco < 0 || endereco >= len(memoria) {
				fmt.Println("Endereço inválido!")
				continue
			}

			switch processador {
			case 1:
				EscreverDado(1, endereco, valor, memoria, &cacheP1, conn, file)
			case 2:
				EscreverDado(2, endereco, valor, memoria, &cacheP2, conn, file)
			case 3:
				EscreverDado(3, endereco, valor, memoria, &cacheP3, conn, file)
			default:
				fmt.Println("Processador inválido!")
				continue
			}

			fmt.Println("Dado escrito com sucesso!")

		case 3:
			// Exibir memória principal
			ExibirMemoriaPrincipal(memoria, conn, file)

		case 4:
			// Exibir caches
			ExibirCache(cacheP1, 1, conn, file)
			ExibirCache(cacheP2, 2, conn, file)
			ExibirCache(cacheP3, 3, conn, file)

		case 5:
			// Sair do programa
			defer conn.Close()
			defer file.Close()
			fmt.Println("Saindo...")
			return

		default:
			fmt.Println("Opção inválida! Tente novamente.")
		}
	}
}
