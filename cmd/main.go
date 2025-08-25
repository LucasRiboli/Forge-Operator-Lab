package main

import (
	"LucasRiboli/operatorLinuxPrimitiveLab/config"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	incus "github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
	"gopkg.in/yaml.v3"
)

const (
	PodPrefix  = "pod"
	ConfigFile = "./hpa.yaml"
)

func main() {
	log.Println("INIT HPA Operator action")

	if err := runHPALoop(); err != nil {
		log.Printf("Erro no HPA: %v", err)
		os.Exit(1)
	}

	log.Println("END HPA Operator action")
}

func getConfig() (config.Config, error) {
	yamlFile, err := os.ReadFile(ConfigFile)
	if err != nil {
		return config.Config{}, fmt.Errorf("erro ao ler config: %w", err)
	}

	var cfg config.Config
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return config.Config{}, fmt.Errorf("erro ao parsear YAML: %w", err)
	}

	return cfg, nil
}

func getIncusClient() (incus.InstanceServer, error) {
	c, err := incus.ConnectIncusUnix("", nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no Incus: %w", err)
	}
	return c, nil
}

func getFullInstances() ([]api.InstanceFull, error) {
	c, err := getIncusClient()
	if err != nil {
		return nil, err
	}

	containers, err := c.GetInstancesFull(api.InstanceTypeContainer)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar containers: %w", err)
	}

	return containers, nil
}

func getInstanceState(podName string) (*api.InstanceState, error) {
	c, err := getIncusClient()
	if err != nil {
		return nil, err
	}

	container, _, err := c.GetInstanceState(podName)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter estado do container %s: %w", podName, err)
	}

	return container, nil
}

func createPod(name string) error {
	c, err := getIncusClient()
	if err != nil {
		return err
	}

	log.Printf("Criando pod: %s", name)

	req := api.InstancesPost{
		Name: name,
		Source: api.InstanceSource{
			Type:  "image",
			Alias: "ubuntu/22.04",
		},
		Type: "container",
	}

	op, err := c.CreateInstance(req)
	if err != nil {
		return fmt.Errorf("erro ao criar container: %w", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("erro na operação de criação: %w", err)
	}

	cmd := exec.Command("incus", "config", "set", name, "limits.memory=1GB")
	err = cmd.Run()
	if err != nil {
		return err
	}

	reqState := api.InstanceStatePut{
		Action:  "start",
		Timeout: -1,
	}

	op, err = c.UpdateInstanceState(name, reqState, "")
	if err != nil {
		return fmt.Errorf("erro ao iniciar container: %w", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("erro na operação de início: %w", err)
	}

	log.Printf("Pod criado e iniciado com sucesso: %s", name)
	return nil
}

func deletePod(name string) error {
	c, err := getIncusClient()
	if err != nil {
		return err
	}

	log.Printf("Deletando pod: %s", name)

	reqState := api.InstanceStatePut{
		Action:  "stop",
		Timeout: 30,
		Force:   true,
	}

	op, err := c.UpdateInstanceState(name, reqState, "")
	if err != nil {
		return fmt.Errorf("erro ao parar container: %w", err)
	}

	op, err = c.DeleteInstance(name)
	if err != nil {
		return fmt.Errorf("erro ao deletar container: %w", err)
	}

	err = op.Wait()
	if err != nil {
		return fmt.Errorf("erro na operação de deleção: %w", err)
	}

	log.Printf("Pod deletado com sucesso: %s", name)
	return nil
}

func getHPAPods(containers []api.InstanceFull) []api.InstanceFull {
	var hpaPods []api.InstanceFull
	for _, container := range containers {
		if strings.HasPrefix(container.Name, PodPrefix) {
			hpaPods = append(hpaPods, container)
		}
	}
	return hpaPods
}

func runHPALoop() error {
	hpa, err := getConfig()
	if err != nil {
		return fmt.Errorf("erro ao carregar config: %w", err)
	}

	containers, err := getFullInstances()
	if err != nil {
		return fmt.Errorf("erro ao listar containers: %w", err)
	}

	hpaPods := getHPAPods(containers)

	var sumMemory int64 = 0
	var runningPods int64 = 0

	for _, container := range hpaPods {
		if container.State.Status == "Running" {
			state, err := getInstanceState(container.Name)
			if err != nil {
				log.Printf("Erro ao capturar estado do pod %s: %v", container.Name, err)
				continue
			}

			if state.Memory.Total > 0 {
				porcentagemPod := (state.Memory.Usage * 100) / state.Memory.Total
				sumMemory += porcentagemPod
				runningPods++
				log.Printf("Pod %s: %d%% memória utilizada", container.Name, porcentagemPod)
			}
		}
	}

	log.Printf("Pods rodando: %d, Threshold: %d%%, Min: %d, Max: %d",
		runningPods, hpa.HPA.MemoryThreshold, hpa.HPA.MinPods, hpa.HPA.MaxPods)

	if runningPods == 0 {
		if hpa.HPA.MinPods > 0 {
			name := PodPrefix + uuid.New().String()[:8]
			return createPod(name)
		}
		return nil
	}

	totalMediaMemory := sumMemory / runningPods
	log.Printf("Média de uso de memoria: %d%%", totalMediaMemory)

	if totalMediaMemory >= int64(hpa.HPA.MemoryThreshold) {
		if runningPods < int64(hpa.HPA.MaxPods) {
			name := PodPrefix + uuid.New().String()[:8]
			log.Printf("Criando novo pod (media: %d%%, threshold: %d%%)",
				totalMediaMemory, hpa.HPA.MemoryThreshold)
			return createPod(name)
		} else {
			log.Printf("já no maximo de pods (%d)", hpa.HPA.MaxPods)
		}
	}

	if totalMediaMemory < int64(hpa.HPA.MemoryThreshold)/2 {
		if runningPods > int64(hpa.HPA.MinPods) {
			podToDelete := hpaPods[len(hpaPods)-1].Name // mudar para pegar o mais antigo
			if podToDelete != "" {
				log.Printf("Deletando pod (média: %d%%, threshold: %d%%)",
					totalMediaMemory, hpa.HPA.MemoryThreshold)
				return deletePod(podToDelete)
			}
		} else {
			log.Printf("ja no minimo de pods (%d)", hpa.HPA.MinPods)
		}
	}

	log.Println("Nenhuma acão de scale")
	return nil
}
