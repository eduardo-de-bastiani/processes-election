package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
    tipo    int    // tipo da mensagem para fazer o controle do que fazer (falha, eleição, confirmação da eleição)
    corpo   [4]int // conteúdo da mensagem para colocar os ids falhos
    lider   int // processo lider da eleição
    started int // processo que iniciou a eleição ou confirmação da eleição
}

var (
    chans = []chan mensagem{ // vetor de canais para formar o anel de eleição - chan[0], chan[1] e chan[2] ...
        make(chan mensagem),
        make(chan mensagem),
        make(chan mensagem),
        make(chan mensagem),
    }
    controle = make(chan int)
    wg       sync.WaitGroup // wg é usado para esperar o programa terminar
)

func ElectionControler(in chan int) {
    defer wg.Done()

    var temp mensagem

	temp.corpo[0] = -1
	temp.corpo[1] = -1
	temp.corpo[2] = -1
	temp.corpo[3] = -1

    // mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isto)
    temp.tipo = 2
    temp.corpo[0] = 0 // falha o processo 0
    chans[0] <- temp	//manda para o 1 (próximo processo no anel)
    fmt.Printf("Controle: mudar o processo 0 para falho\n")
	
    fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

    // mudar o processo 1 - canal de entrada 0 - para falho (defini mensagem tipo 2 pra isto)
    temp.tipo = 2
	temp.corpo[1] = 1 // falha o processo 1
    chans[1] <- temp //manda para o 2 (próximo processo no anel)
    fmt.Printf("Controle: mudar o processo 1 para falho\n")
    fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação


    fmt.Println("\n   Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
    defer wg.Done()

    var FailedId1 int
    var FailedId2 int
    var FailedId3 int
    var FailedId4 int

    for {
        temp := <-in // ler mensagem
        FailedId1 = temp.corpo[0]
        FailedId2 = temp.corpo[1]
        FailedId3 = temp.corpo[2]
        FailedId4 = temp.corpo[3]
        fmt.Printf("%2d: recebi mensagem do tipo %d, [ %d, %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])

        switch temp.tipo {
        case 2:
            {
                fmt.Printf("%2d: disparando eleição\n", TaskId)
                fmt.Printf("%2d: processos falhos [%d, %d, %d, %d]\n", TaskId, FailedId1, FailedId2, FailedId3, FailedId4)
				controle <- -5

                // vai pra eleição
                temp.tipo = 0
                temp.started = TaskId
                temp.lider = TaskId
                fmt.Printf("%2d: líder da eleição: %d\n", TaskId, temp.lider)
                out <- temp
            }
        case 0:
            {
                // se o processo está falho, apenas repassa a msg
                if TaskId == FailedId1 || TaskId == FailedId2 || TaskId == FailedId3 || TaskId == FailedId4 {
                    fmt.Printf("%2d: processo falho, repassando mensagem...\n", TaskId)
                    out <- temp
                    break
                }
                // se a msg chega no processo que iniciou a eleição, chamamos a confirmação
                if TaskId == temp.started {
                    fmt.Printf("%2d: encerrando votação\n", TaskId)
                    fmt.Printf("%2d: novo líder é %d\n", TaskId, temp.lider)
                    temp.tipo = 1
                    out <- temp
                    break
                }
                // votação
                if TaskId != temp.started {
                    fmt.Printf("%2d: votação\n", TaskId)
                    temp.tipo = 0
                    if TaskId < temp.lider {
                        fmt.Printf("%2d: sou menor que o líder atual (%d)\n", TaskId, temp.lider)
                        temp.lider = TaskId
						fmt.Printf("%2d:líder da votação: %d\n", TaskId, temp.lider)
                    } else {
                        fmt.Printf("%2d: sou maior que o líder atual (%d), passando a mensagem...\n", TaskId, temp.lider)
                    }
                    out <- temp
                }
            }
        case 1:
            {
                // se o processo está falho, apenas repassa a msg
                if TaskId == FailedId1 || TaskId == FailedId2 || TaskId == FailedId3 || TaskId == FailedId4 {
                    fmt.Printf("%2d: processo falho, repassando mensagem...\n", TaskId)
                    out <- temp
                    break
                }
                // se a msg chega no processo que iniciou a confirmação, encerramos a eleição
                if TaskId == temp.started {
                    fmt.Printf("%2d: encerrando eleição\n", TaskId)
                    controle <- -5
                    break
                }
                // atualização de lider para cada processo
                if TaskId != temp.started {
                    fmt.Printf("%2d: atualizando meu líder para %d\n", TaskId, temp.lider)
                    temp.tipo = 1
                    out <- temp
                }
            }

        default:
            {
                fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
            }
        }
        fmt.Printf("%2d: terminei \n", TaskId)

    }
}

func main() {

    wg.Add(5) // Adiciona uma contagem de cinco, uma para cada goroutine

    fmt.Println("\n   Processo controlador criado\n")
    go ElectionControler(controle)

    fmt.Println("\n   Anel de processos criado\n")
	
    go ElectionStage(0, chans[3], chans[0], 0) // este é o líder
    go ElectionStage(1, chans[0], chans[1], 0) // não é líder, é o processo 0
    go ElectionStage(2, chans[1], chans[2], 0) // não é líder, é o processo 0
    go ElectionStage(3, chans[2], chans[3], 0) // não é líder, é o processo 0
	
	
    wg.Wait() // Espera as goroutines terminarem
}