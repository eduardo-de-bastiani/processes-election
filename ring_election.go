package main

import (
	"fmt"
	"sync"
)

var wg sync.WaitGroup

type mensagem struct {
	tipo  int    // 0: eleição, 1: coordenador
	corpo int // conteúdo da mensagem para colocar os ids (usar um tamanho compatível com o número de processos no anel)
	lider int    // PID do processo que será o coordenador
}

var chans [4]chan mensagem

func ElectionControler(in chan int) {
	defer wg.Done()
	var temp mensagem
	// Comandos para falhar os processos
	temp.tipo = 2 // Tipo para falha
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação
	// Mudar o processo 1 para falho
	// temp.tipo = 2
	// chans[0] <- temp
	// fmt.Printf("Controle: mudar o processo 1 para falho\n")
	// fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação
	// fmt.Println("\n Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int, control chan int, initiator int) {
	defer wg.Done()

	var actualLeader int
    var bFailed bool = false
	actualLeader = leader
	localInitiator := initiator // Keep track of initiator locally
    actualLeader = leader

	for {
		msg := <-in // Recebe a mensagem do processo anterior

		fmt.Printf("%2d: recebi mensagem  tipo %d, lider: %d\n", TaskId, msg.tipo, msg.lider)



		switch msg.tipo {
		case 2: // Processo falhou
			bFailed = true
			fmt.Printf("%2d: processo líder %d falho\n", TaskId, leader)
			// Aqui, o processo anterior deve iniciar a eleição
			// Iniciar eleição
            newMsg := mensagem{
                tipo: 0,
                corpo: TaskId,
                lider: -1,
            }
			// Enviar mensagem de eleição
			control <- TaskId
			localInitiator = TaskId // Set this process as initiator
            out <- newMsg
            continue

		case 0: // Mensagem de eleição
			if bFailed{
				out <- msg
				continue
			}
			if localInitiator == TaskId {
                // Transformar em mensagem de líder
                newMsg := mensagem{
                    tipo: 1,
                    corpo: TaskId,
                    lider: msg.corpo, // O líder será o processo com menor ID
                }
				localInitiator = -1 // Reset initiator
				out <- newMsg
                continue

            }

			if localInitiator == -1 {
				localInitiator = msg.corpo // Guarda quem iniciou a eleição
				fmt.Printf("%2d: Registrando initiator como %d\n", TaskId, localInitiator)
			}

			// Compara IDs e atualiza o corpo se necessário
			if TaskId < msg.corpo {
				msg.corpo = TaskId
				fmt.Printf("%2d: Atualizando corpo para %d\n", TaskId, TaskId)
			}


			out <- msg

		case 1: // mensagem de coordenador
			bFailed = false
			actualLeader = msg.lider
			fmt.Printf("%2d: novo líder é %d\n", TaskId, actualLeader)


			 // Repassa a mensagem se não completou o ciclo
			 if msg.corpo == TaskId {
                fmt.Printf("%2d: eleição concluída\n", TaskId)
                return // Encerra o processo
            }

			out <- msg
		default:
			fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
		}

		fmt.Printf("%2d: terminei \n", TaskId)
	}
}

func main() {
	in := make(chan int)

	for i := 0; i < 4; i++{
		chans[i] = make(chan mensagem)
	}

	wg.Add(1)
	go ElectionControler(in)

	for i:= 0; i < 4; i++{
		wg.Add(1)
		go ElectionStage(i, chans[i], chans[(i+1)%4], 0, in, -1)
	}

	wg.Wait()
}
