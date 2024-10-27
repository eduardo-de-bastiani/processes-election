// Código exemplo para o trabaho de sistemas distribuidos (eleicao em anel)
// By Cesar De Rose - 2022

package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo [4]int // conteudo da mensagem para colocar os ids (usar um tamanho ocmpativel com o numero de processos no anel)
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

	// comandos para o anel iciam aqui

	// mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isto)

	temp.tipo = 2
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")

	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// mudar o processo 1 - canal de entrada 0 - para falho (defini mensagem tipo 2 pra isto)

	// temp.tipo = 2
	// chans[0] <- temp
	// fmt.Printf("Controle: mudar o processo 1 para falho\n")
	// fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// matar os outrs processos com mensagens não conhecidas (só pra cosumir a leitura)

	// temp.tipo = 4
	// chans[1] <- temp
	// chans[2] <- temp

	fmt.Println("\n   Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	// variaveis locais que indicam se este processo é o lider e se esta ativo

	// var actualLeader int
	var bFailed bool = false // todos inciam sem falha

	//actualLeader = leader // indicação do lider veio por parâmatro
	var newLeader = -1
	var voted = 0
	var count = 0

	temp := <-in // ler mensagem
	fmt.Printf("%2d: recebi mensagem do tipo %d, [ %d, %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])

	switch temp.tipo {
	case 2:
		{
			bFailed = true
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			controle <- -5

			// vai pra eleição
			temp.tipo = 0
			temp.corpo[0] = -1
			temp.corpo[1] = -1
			temp.corpo[2] = -1
			temp.corpo[3] = -1
			out <- temp
		}
	case 0:
		{

			// para cada processo no anel, se o processo não estiver falho, ele coloca seu TaskId no corpo da mensagem
			// e envia a mensagem para o próximo processo no anel
			if !bFailed {
				fmt.Printf("%2d: votação\n", TaskId)
				if temp.corpo[0] == -1 {
					temp.corpo[0] = TaskId
					} else if temp.corpo[1] == -1 {
						temp.corpo[1] = TaskId
						} else if temp.corpo[2] == -1 {
							temp.corpo[2] = TaskId
						}
						fmt.Printf("%2d: passando mensagem do tipo %d, [ %d, %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])
					}
					voted++

			if count == 4{
				temp.tipo = 1
			}
			out <- temp


		}
	case 1:
		{
			fmt.Printf("%2d: Apuração dos votos\n", TaskId)

			if count == 4{
				break;
			}
			// para cada processo no anel, se o processo for menor que o lider atual, ele atualiza o lider
			for i := 0; i < 4; i++ {
				if TaskId < temp.corpo[i] {
					newLeader = TaskId
				}
				count++
				out <- temp
			}

			if temp.corpo[0] == TaskId {
				fmt.Printf("%2d: novo lider é %d\n", TaskId, newLeader)
			}
		}

	default:
		{
			fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)

		}

		fmt.Printf("%2d: terminei \n", TaskId)
	}
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
