# processes-election

## Modificações

Ao invés da mensagem conter todos os PIDs dos processos que votaram e depois serem comparados para se definir o novo coordenador, o menor PID se mantém na mensagem durante a eleição, só sendo trocado se algum processo obtiver um PID menor. Quando a mensagem chegar no processo que disparou a eleição, o novo coordenador é automaticamente anunciado.