package orchestrator

import (
	kubernetes "k8s.io/client-go/kubernetes"
)

type Orchestrator struct {
	Client  *kubernetes.Clientset
	CSImage string
}

func NewOrchestrator(client *kubernetes.Clientset, csImage string) *Orchestrator {
	return &Orchestrator{Client: client, CSImage: csImage}
}
