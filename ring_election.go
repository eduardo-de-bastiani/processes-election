package main

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

type mensagem struct {
	tipo  int    // 0: eleição, 1: coordenador
	corpo [4]int // conteúdo da mensagem para colocar os ids (usar um tamanho compatível com o número de processos no anel)
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
	time.Sleep(2 * time.Second)
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

		fmt.Printf("%2d: recebi mensagem  tipo %d, lider: %d\n", TaskId, msg.tipo, msg.lider)



		switch msg.tipo {
		case 2: // Processo falhou
			bFailed = true
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			// Aqui, o processo anterior deve iniciar a eleição
			// Enviar mensagem de eleição
			msg.tipo = 0 // Tipo de mensagem de eleição
			msg.corpo[TaskId] = TaskId
			//println("Teste lider depois da falha: ", msg.lider)
			out <- msg // repassa a mensagem para o próximo processo
			continue   // continua no loop

		case 0: // Mensagem de eleição
			if !bFailed{
				msg.corpo[TaskId] = TaskId
			}
			// Se a mensagem voltou ao processo que iniciou a eleição
			if msg.corpo[TaskId] == TaskId {
				// Determina o novo líder (o menor PID)
				// se a mensagem chegou no processo que falhou, ele não pode ser líder e deve ser ignorado

				newLeader := TaskId
				for _, pid := range msg.corpo {
					if pid < newLeader {
						newLeader = pid
					}
				}
			msg.tipo = 1 // Tipo de mensagem de coordenador
			msg.lider = newLeader
			
		}
			// Repassa a mensagem para o próximo processo
			out <- msg
			continue // continua no loop

		case 1: // mensagem de coordenador
			bFailed = false
			actualLeader = msg.lider
			fmt.Printf("%2d: novo líder é %d\n", TaskId, actualLeader)

			// Atualiza seu líder e repassa para os outros processos
			fmt.Printf("%2d: atualizando líder para %d\n", TaskId, actualLeader)
			msg.lider = actualLeader
			// Repassa a mensagem para o próximo processo no anel
			if msg.corpo[TaskId] != TaskId {
				out <- msg
			}
			break // sai do loop após atualizar e repassar

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
		go ElectionStage(i, chans[i], chans[(i+1)%4], 0)
	}

	wg.Wait()
}
