# Operador HPA com Primitivas Linux

Este projeto é uma prova prática do "Operator Pattern", demonstrando como a lógica de um operador — que monitora e reconcilia o estado de recursos — pode ser implementada fora do Kubernetes, usando ferramentas básicas do Linux.

### O que o Operador em questão faz?

Ele atua como um HPA de pods Incus, baseando suas decisões no uso de memória. 

O operador:
* Monitora o uso de memória de containers com o prefixo `pod`.
* Escala o número de pods para cima ou para baixo para manter a utilização de memória abaixo de um limite definido.

### Tecnologias Usadas
* **Go**: Linguagem de programação do operador.
* **Incus**: Para gerenciar os containers.
* **systemd**: Para rodar o operador como um serviço daemon.
* **Shell Script**: Para a instalação e desinstalação.

### Como Usar

1.  **Prepare o ambiente**:
    Execute o `lab-init.sh` para criar o container inicial do Incus.

2.  **Compile o operador**:
    Rode `go build -o operador cmd/main.go`.

3.  **Instale e inicie o serviço**:
    Execute `sudo ./operador/install-daemon.sh` para instalar o operador como um serviço `systemd`.
    Em seguida, rode `sudo systemctl start operator` para iniciá-lo.

4.  **Configure os limites**:
    Edite o arquivo `hpa.yaml` para ajustar os parâmetros de auto-escalonamento.

### Exemplo de Configuração (`hpa.yaml`)
```yaml
HPA:
  minPods: 1
  maxPods: 5
  cpuThreshold: 80
  memoryThreshold: 70