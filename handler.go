package main

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
)

func restartsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request for %s", r.URL.Path)

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podname, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var restartCount int32

	for _, containerStatus := range pod.Status.ContainerStatuses {
		restartCount += containerStatus.RestartCount
	}

	data := map[string]int32{
		"restarts": restartCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func cpuInfoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request for %s", r.URL.Path)
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podname, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Assuming the first container is the one we're interested in
	if len(pod.Spec.Containers) > 0 {
		resources := pod.Spec.Containers[0].Resources
		cpuLimit := resources.Limits.Cpu().String()
		cpuRequest := resources.Requests.Cpu().String()

		data := map[string]string{
			"limit":   cpuLimit,
			"request": cpuRequest,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	} else {
		log.Println("No containers found in deployment")
		http.Error(w, "No containers found in deployment", http.StatusInternalServerError)
	}
}

func memInfoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request for %s", r.URL.Path)

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podname, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(pod.Spec.Containers) > 0 {
		resources := pod.Spec.Containers[0].Resources
		memLimit := resources.Limits.Memory().String()
		memRequest := resources.Requests.Memory().String()

		data := map[string]string{
			"limit":   memLimit,
			"request": memRequest,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	} else {
		log.Println("No containers found in deployment")
		http.Error(w, "No containers found in deployment", http.StatusInternalServerError)
	}
}

type PatchData struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

func patchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request for %s", r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var data PatchData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Error decoding request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("%+v\n", data)

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podname, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(pod.Spec.Containers) == 0 {
		http.Error(w, "No containers found in deployment", http.StatusInternalServerError)
		return
	}

	container := &pod.Spec.Containers[0]
	res := container.Resources
	if res.Requests == nil {
		res.Requests = make(corev1.ResourceList)
	}
	if res.Limits == nil {
		res.Limits = make(corev1.ResourceList)
	}

	// Update CPU
	if data.CPU != "" {
		cpu, err := resource.ParseQuantity(data.CPU)
		if err != nil {
			http.Error(w, "Invalid CPU request value: "+err.Error(), http.StatusBadRequest)
			return
		}
		res.Requests[corev1.ResourceCPU] = cpu
		res.Limits[corev1.ResourceCPU] = cpu
	}

	// Update Memory
	if data.Memory != "" {
		memory, err := resource.ParseQuantity(data.Memory)
		if err != nil {
			http.Error(w, "Invalid Memory request value: "+err.Error(), http.StatusBadRequest)
			return
		}
		res.Requests[corev1.ResourceMemory] = memory
		res.Limits[corev1.ResourceMemory] = memory
	}

	patch := map[string]any{
		"spec": map[string]any{
			"containers": []map[string]any{
				// add container in this array
				{
					"name":      container.Name,
					"resources": res,
				},
			},
		},
	}

	patchData, err := json.Marshal(patch)
	if err != nil {
		http.Error(w, "Data is not valid JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("%s\n", patchData)

	_, err = clientset.CoreV1().Pods(namespace).Patch(context.TODO(), pod.Name, k8sTypes.StrategicMergePatchType, patchData, metav1.PatchOptions{}, "resize")
	if err != nil {
		http.Error(w, "Failed to update pod: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Failed to update pod: %v", err)
		return
	}

	log.Printf("Deployment %s patched successfully with new resource values\n", podname)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Deployment patched successfully"})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request for %s", r.URL.Path)
	tmpl, err := template.ParseFS(templates, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
