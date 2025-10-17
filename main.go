package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset
var podname string
var namespace string

//go:embed templates
var templates embed.FS

func initKubeClient() error {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		log.Printf("Application is running inside a Kubernetes cluster.")
		var err error
		clientset, err = getKubernetesClient()
		if err != nil {
			return fmt.Errorf("failed to get Kubernetes client: %w", err)
		}
	} else {
		log.Printf("Application is NOT running inside a Kubernetes cluster.")
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

// getKubernetesClient sets up the Kubernetes client configuration.
// It prioritizes in-cluster configuration (for running inside a Pod).
// If in-cluster config fails, it will attempt to use the local kubeconfig file
// (useful for local development/testing outside the cluster).
func getKubernetesClient() (*kubernetes.Clientset, error) {
	// Try to get in-cluster config (standard for Pods)
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}
	return clientset, nil
}

func main() {
	if err := initKubeClient(); err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	namespace = os.Getenv("NAMESPACE")
	podname = os.Getenv("POD_NAME")

	if namespace == "" || podname == "" {
		namespace = "ippr"
		podname = "ippr"
	}

	http.HandleFunc("/", loggingMiddleware(http.HandlerFunc(homeHandler)))
	http.HandleFunc("/api/cpuInfo", loggingMiddleware(http.HandlerFunc(cpuInfoHandler)))
	http.HandleFunc("/api/memInfo", loggingMiddleware(http.HandlerFunc(memInfoHandler)))
	http.HandleFunc("/api/restarts", loggingMiddleware(http.HandlerFunc(restartsHandler)))
	http.HandleFunc("/api/patch", loggingMiddleware(http.HandlerFunc(patchHandler)))
	log.Printf("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

/*

kubectl patch pod -n ippr ippr-deployment-67f7c69b5c-zbwr2 --subresource resize --patch \
  '{"spec":{"containers":[{"name":"ippr", "resources":{"requests":{"memory":"1200Mi"}, "limits":{"memory":"1200Mi"}}}]}}'

*/
