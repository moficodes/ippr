package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset
var deployment string
var namespace string

//go:embed templates
var templates embed.FS

func initKubeClient() error {
	kubeconfig, err := getKubeConfig()
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build config from flags: %w", err)
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}
	return nil
}

func getKubeConfig() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".kube", "config"), nil
}

func main() {
	if err := initKubeClient(); err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	deployment = os.Getenv("DEPLOYMENT_NAME")
	namespace = os.Getenv("NAMESPACE")

	if deployment == "" || namespace == "" {
		deployment = "ippr-deployment"
		namespace = "ippr"
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/cpuInfo", cpuInfoHandler)
	http.HandleFunc("/api/memInfo", memInfoHandler)
	http.HandleFunc("/api/restarts", restartsHandler)
	http.HandleFunc("/api/patch", patchHandler)
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
