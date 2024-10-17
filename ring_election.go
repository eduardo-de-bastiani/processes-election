package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // 0: eleição, 1: coordenador
	corpo [3]int // conteúdo da mensagem para colocar os ids (usar um tamanho compatível com o número de processos no anel)
	lider int    // PID do processo que será o coordenador
}

var (
	chans = []chan mensagem{ // vetor de canais para formar o anel de eleição
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle = make(chan int)
	wg       sync.WaitGroup // wg is used to wait for the program to finish
)

func ElectionControler(in chan int) {
	defer wg.Done()
	var temp mensagem
	// Comandos para falhar os processos
	temp.tipo = 2 // Tipo para falha
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação
	// Mudar o processo 1 para falho
	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mudar o processo 1 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação
	fmt.Println("\n Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	var actualLeader int
    var bFailed bool = false
	actualLeader = leader

	for {
		msg := <-in // Recebe a mensagem do processo anterior

		fmt.Printf("%2d: recebi mensagem %d, lider: %d\n", TaskId, msg.tipo, msg.lider)

		switch msg.tipo {
		case 2: // Processo falhou
			bFailed = true
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
			// Aqui, o processo anterior deve iniciar a eleição
			// Enviar mensagem de eleição
			msg.tipo = 1 // Tipo de mensagem de eleição
			msg.lider = TaskId
			//println("Teste lider depois da falha: ", msg.lider)
			out <- msg // repassa a mensagem para o próximo processo
			continue   // continua no loop

		case 1: // Mensagem de eleição
			// Atualiza o PID se for menor que o atual
			if msg.lider < actualLeader {
				actualLeader = msg.lider
				fmt.Printf("%2d: voto atual %d\n", TaskId, actualLeader)
			}
			// Repassa a mensagem para o próximo processo
			out <- msg
			continue // continua no loop

		case 3: // Confirmação de eleição
			bFailed = false
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
			continue // continua no loop

		default:
			fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
		}

		// Verifica se é o processo que iniciou a eleição
		if msg.lider == TaskId {
			// Atualiza seu líder e repassa para os outros processos
			fmt.Printf("%2d: atualizando líder para %d\n", TaskId, actualLeader)
			msg.lider = actualLeader
			// Repassa a mensagem para o próximo processo no anel
			out <- msg
			break // sai do loop após atualizar e repassar
		}

		fmt.Printf("%2d: terminei \n", TaskId)
	}
}

func main() {
	// Cria os canais para comunicação entre processos
	wg.Add(5) // Adiciona um contador de quatro, um para cada goroutine
    // Criar os processos do anel de eleição
    go ElectionStage(0, chans[3], chans[0], 0) // este é o líder
    go ElectionStage(1, chans[0], chans[1], 0) // não é líder
    go ElectionStage(2, chans[1], chans[2], 0) // não é líder
    go ElectionStage(3, chans[2], chans[3], 0) // não é líder


	 fmt.Println("\n Anel de processos criado")
    // Criar o processo controlador
    go ElectionControler(controle)
    fmt.Println("\n Processo controlador criado\n")

    wg.Wait() // Aguarda as goroutines terminarem
}
