// Código exemplo para o trabaho de sistemas distribuidos (eleicao em anel)
// By Cesar De Rose - 2022

package main

import (
	"fmt"
	"sync"
	"time"
)

const (
    MSG_NORMAL    = 1  // mensagem normal			// REMOVER
    MSG_FAIL      = 2  // Process failure message
    MSG_RECOVER   = 3  // Process recovery message
    MSG_ELECTION  = 4  // Start election message
    MSG_ELECTED   = 5  // New coordinator announcement
    MSG_PING      = 6  // Ping message to check coordinator
    MSG_PONG      = 7  // Coordinator response to ping
)



type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo [3]int // conteudo da mensagem para colocar os ids (usar um tamanho ocmpativel com o numero de processos no anel)
	recebedor int
}

var (
	chans = []chan mensagem{ // vetor de canias para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
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

	// estabilização do sistema
	time.Sleep(time.Second * 2)
	// comandos para o anel iciam aqui


	// mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isto)

	temp.tipo = MSG_FAIL
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação


	time.Sleep(time.Second * 5)

	// mudar o processo 1 - canal de entrada 0 - para falho (defini mensagem tipo 2 pra isto)

	temp.tipo = MSG_FAIL
	chans[0] <- temp
	fmt.Printf("Controle: mudar o processo 1 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// timesleep para observar os processos falhando e elegendo novos coordenadores
	time.Sleep(time.Second * 30)
	fmt.Println("\n   Processo controlador concluído\n")
}


func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
    defer wg.Done()
    // variaveis locais que indicam se este processo é o lider e se esta ativo
    var actualLeader int = leader
    var bFailed bool = false // todos inciam sem falha
    timeout := time.Second * 3  // Added := for declaration

    if TaskId != actualLeader {
        go monitorCoordinator(TaskId, actualLeader, in, out)
    }

    for {
        select {
        case temp := <-in: {  // ler mensagem
            fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])
            switch temp.tipo {
            case MSG_FAIL:
                bFailed = true
                fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
                fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
                controle <- -5

            case MSG_RECOVER:
                bFailed = false
                fmt.Printf("%2d: recuperou %v \n", TaskId, bFailed)
                fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
                controle <- -5

            case MSG_ELECTION:
                if bFailed {
                    // Processos falhos não participam
                    continue
                }

                if temp.corpo[0] == TaskId {
                    // Eleição deu uma volta completa
                    announceNewCoordinator(TaskId, temp.corpo[2], out)
                    continue
                }

                // verificação para eleger (menor PID)
                proposedCoordinator := temp.corpo[2]
                if TaskId < proposedCoordinator {
                    proposedCoordinator = TaskId
                }

                // propaga a mensagem de eleição
                out <- mensagem{
                    tipo:  MSG_ELECTION,
                    corpo: [3]int{temp.corpo[0], temp.corpo[1], proposedCoordinator},
                }

            case MSG_ELECTED:
                if !bFailed {
                    actualLeader = temp.corpo[2]
                    fmt.Printf("%2d: novo cordenador: %d\n", TaskId, actualLeader)
                    // propaga o resultado da eleição (novo coordenador)
                    if temp.corpo[0] != TaskId {
                        out <- temp
                    }
                }
            }
        }

        case <-time.After(timeout): {
            if bFailed || TaskId == actualLeader {
                continue
            }
            // timeout, possivelmente coordenador falhou
            fmt.Printf("%2d: timeout detectou coordenador %d, iniciando eleição\n", TaskId, actualLeader)
            startElection(TaskId, out)
        }
        }
    }
}
func monitorCoordinator(TaskId int, coordinator int, in chan mensagem, out chan mensagem) {
	for {
		time.Sleep(time.Second) // espera entre os pings
		out <- mensagem{
			tipo:     MSG_PING,
			corpo:    [3]int{TaskId, coordinator, 0},
			recebedor: coordinator,
		}
	}
}

func startElection(TaskId int, out chan mensagem) {
    // Inicia eleição votando em si mesmo como coordenador
    out <- mensagem{
        tipo:  MSG_ELECTION,
        corpo: [3]int{TaskId, TaskId, TaskId}, // [election_starter, current_pid, proposed_coordinator]
    }
}

func announceNewCoordinator(TaskId int, coordinator int, out chan mensagem) {
    out <- mensagem{
        tipo:  MSG_ELECTED,
        corpo: [3]int{TaskId, 0, coordinator},
    }
    fmt.Printf("%2d: announcing new coordinator %d\n", TaskId, coordinator)
}



func main() {

	wg.Add(5) // Add a count of four, one for each goroutine

	// criar os processo do anel de eleicao

	go ElectionStage(0, chans[3], chans[0], 0) // este é o lider
	go ElectionStage(1, chans[0], chans[1], 0) // não é lider, é o processo 0
	go ElectionStage(2, chans[1], chans[2], 0) // não é lider, é o processo 0
	go ElectionStage(3, chans[2], chans[3], 0) // não é lider, é o processo 0

	fmt.Println("\n   Anel de processos criado")

	// criar o processo controlador

	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	wg.Wait() // Wait for the goroutines to finish\
}
