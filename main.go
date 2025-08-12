package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const asciiLogo = `
                      .-')    
                     ( OO ).  
     ,--.  .-----.  (_)---\_) 
 .-')| ,| /  .-.  \ /    _ |  
( OO |(_||   \_.' / \  :´ ´.  
| ´-'|  | /  .-. '.  ´..'''.) 
,--. |  ||  |   |  |.-._)   \ 
|  '-'  / \  '-'  / \       / 
 ´-----'   ´----''   ´-----'  
`

// Pod represents a Kubernetes pod from kubectl output
type Pod struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Containers []Container `json:"containers"`
	} `json:"spec"`
	Status struct {
		Phase string `json:"phase"`
	} `json:"status"`
}

// Container represents a container within a pod
type Container struct {
	Name string `json:"name"`
}

// PodList represents the kubectl get pods output
type PodList struct {
	Items []Pod `json:"items"`
}

func printLogo() {
	fmt.Print(asciiLogo)
	fmt.Println("j8s is a Java CLI tool for Kubernetes")
	fmt.Println()
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  j8s jstack    Interactive JVM thread dump from Kubernetes pod")
	fmt.Println("  j8s dumpheap  Interactive JVM heap dump from Kubernetes pod (downloads locally)")
	fmt.Println("  j8s           Show this help")
	fmt.Println()
}

func main() {
	printLogo()
	
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "jstack":
		if err := runJstack(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "dumpheap":
		if err := runDumpheap(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func runJstack() error {
	// Get list of pods
	pods, err := listPods()
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	if len(pods) == 0 {
		return fmt.Errorf("no pods found in current namespace")
	}

	// Select pod
	selectedPod, err := selectPod(pods)
	if err != nil {
		return fmt.Errorf("failed to select pod: %v", err)
	}

	// Select container if multiple
	var selectedContainer string
	if len(selectedPod.Spec.Containers) > 1 {
		selectedContainer, err = selectContainer(selectedPod.Spec.Containers)
		if err != nil {
			return fmt.Errorf("failed to select container: %v", err)
		}
	} else {
		selectedContainer = selectedPod.Spec.Containers[0].Name
	}

	fmt.Printf("Selected pod: %s, container: %s\n", selectedPod.Metadata.Name, selectedContainer)
	
	// Deploy jattach to the container
	if err := deployJattach(selectedPod.Metadata.Name, selectedContainer); err != nil {
		return fmt.Errorf("failed to deploy jattach: %v", err)
	}

	// Find Java process and run thread dump
	if err := runThreadDump(selectedPod.Metadata.Name, selectedContainer); err != nil {
		return fmt.Errorf("failed to run thread dump: %v", err)
	}

	return nil
}

func runDumpheap() error {
	// Get list of pods
	pods, err := listPods()
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	if len(pods) == 0 {
		return fmt.Errorf("no pods found in current namespace")
	}

	// Select pod
	selectedPod, err := selectPod(pods)
	if err != nil {
		return fmt.Errorf("failed to select pod: %v", err)
	}

	// Select container if multiple
	var selectedContainer string
	if len(selectedPod.Spec.Containers) > 1 {
		selectedContainer, err = selectContainer(selectedPod.Spec.Containers)
		if err != nil {
			return fmt.Errorf("failed to select container: %v", err)
		}
	} else {
		selectedContainer = selectedPod.Spec.Containers[0].Name
	}

	fmt.Printf("Selected pod: %s, container: %s\n", selectedPod.Metadata.Name, selectedContainer)
	
	// Deploy jattach to the container
	if err := deployJattach(selectedPod.Metadata.Name, selectedContainer); err != nil {
		return fmt.Errorf("failed to deploy jattach: %v", err)
	}

	// Find Java process and run heap dump
	if err := runHeapDump(selectedPod.Metadata.Name, selectedContainer); err != nil {
		return fmt.Errorf("failed to run heap dump: %v", err)
	}

	return nil
}

func listPods() ([]Pod, error) {
	cmd := exec.Command("kubectl", "get", "pods", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("kubectl command failed: %v", err)
	}

	var podList PodList
	if err := json.Unmarshal(output, &podList); err != nil {
		return nil, fmt.Errorf("failed to parse kubectl output: %v", err)
	}

	// Filter running pods
	var runningPods []Pod
	for _, pod := range podList.Items {
		if pod.Status.Phase == "Running" {
			runningPods = append(runningPods, pod)
		}
	}

	return runningPods, nil
}

func selectPod(pods []Pod) (Pod, error) {
	if len(pods) == 1 {
		fmt.Printf("Using pod: %s\n", pods[0].Metadata.Name)
		return pods[0], nil
	}

	fmt.Println("Available pods:")
	for i, pod := range pods {
		fmt.Printf("%d. %s\n", i+1, pod.Metadata.Name)
	}

	fmt.Print("Select pod (1-" + fmt.Sprintf("%d", len(pods)) + "): ")
	var selection string
	if _, err := fmt.Scanln(&selection); err != nil {
		return Pod{}, fmt.Errorf("failed to read selection: %v", err)
	}

	selection = strings.TrimSpace(selection)
	if selection == "" {
		return Pod{}, fmt.Errorf("no selection made")
	}

	// Parse selection
	var index int
	if _, err := fmt.Sscanf(selection, "%d", &index); err != nil {
		return Pod{}, fmt.Errorf("invalid selection: %s", selection)
	}

	if index < 1 || index > len(pods) {
		return Pod{}, fmt.Errorf("selection out of range: %d", index)
	}

	return pods[index-1], nil
}

func selectContainer(containers []Container) (string, error) {
	fmt.Println("Available containers:")
	for i, container := range containers {
		fmt.Printf("%d. %s\n", i+1, container.Name)
	}

	fmt.Print("Select container (1-" + fmt.Sprintf("%d", len(containers)) + "): ")
	var selection string
	if _, err := fmt.Scanln(&selection); err != nil {
		return "", fmt.Errorf("failed to read selection: %v", err)
	}

	selection = strings.TrimSpace(selection)
	if selection == "" {
		return "", fmt.Errorf("no selection made")
	}

	// Parse selection
	var index int
	if _, err := fmt.Sscanf(selection, "%d", &index); err != nil {
		return "", fmt.Errorf("invalid selection: %s", selection)
	}

	if index < 1 || index > len(containers) {
		return "", fmt.Errorf("selection out of range: %d", index)
	}

	return containers[index-1].Name, nil
}

func deployJattach(podName, containerName string) error {
	fmt.Println("Deploying jattach to container...")

	// First try to copy locally downloaded jattach
	if err := copyJattachToContainer(podName, containerName); err != nil {
		fmt.Printf("kubectl cp failed (%v), trying in-container download...\n", err)
		// Fallback to in-container download
		if err := downloadJattachInContainer(podName, containerName); err != nil {
			return fmt.Errorf("both kubectl cp and in-container download failed: %v", err)
		}
	}

	// Make jattach executable
	fmt.Println("Making jattach executable...")
	cmd := exec.Command("kubectl", "exec", podName, "-c", containerName, "--", "chmod", "+x", "/tmp/jattach")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to make jattach executable: %v", err)
	}

	return nil
}

func copyJattachToContainer(podName, containerName string) error {
	// Download jattach locally first
	localPath, err := downloadJattachLocally()
	if err != nil {
		return fmt.Errorf("failed to download jattach locally: %v", err)
	}
	defer os.Remove(localPath) // Clean up

	// Copy to container
	cmd := exec.Command("kubectl", "cp", localPath, fmt.Sprintf("%s:/tmp/jattach", podName), "-c", containerName)
	return cmd.Run()
}

func downloadJattachLocally() (string, error) {
	// Create temporary file
	tmpDir := os.TempDir()
	localPath := filepath.Join(tmpDir, "jattach")

	// Download jattach binary - try the main repository
	url := "https://github.com/jattach/jattach/releases/latest/download/jattach"
	
	fmt.Printf("Downloading jattach from %s...\n", url)
	
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download jattach: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download jattach: status %d", resp.StatusCode)
	}

	// Create the file
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %v", err)
	}
	defer file.Close()

	// Copy the content
	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save jattach: %v", err)
	}

	// Make it executable
	if err := os.Chmod(localPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make jattach executable: %v", err)
	}

	return localPath, nil
}

func downloadJattachInContainer(podName, containerName string) error {
	url := "https://github.com/jattach/jattach/releases/latest/download/jattach"
	
	// Try curl first
	fmt.Println("Attempting download with curl...")
	cmd := exec.Command("kubectl", "exec", podName, "-c", containerName, "--", "curl", "-L", "-o", "/tmp/jattach", url)
	if err := cmd.Run(); err != nil {
		// Try wget as fallback
		fmt.Println("curl failed, trying wget...")
		cmd = exec.Command("kubectl", "exec", podName, "-c", containerName, "--", "wget", "-O", "/tmp/jattach", url)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("both curl and wget failed: %v", err)
		}
	}

	return nil
}

func runThreadDump(podName, containerName string) error {
	fmt.Println("Finding Java process...")
	
	// Find Java PID using pidof first, then pgrep as fallback
	var javaPID string
	
	// Try pidof java
	cmd := exec.Command("kubectl", "exec", podName, "-c", containerName, "--", "pidof", "java")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		javaPID = strings.TrimSpace(string(output))
		// If multiple PIDs, take the first one
		if strings.Contains(javaPID, " ") {
			javaPID = strings.Split(javaPID, " ")[0]
		}
	} else {
		// Try pgrep as fallback
		cmd = exec.Command("kubectl", "exec", podName, "-c", containerName, "--", "pgrep", "-f", "java")
		output, err = cmd.Output()
		if err != nil || len(output) == 0 {
			return fmt.Errorf("no Java process found in container")
		}
		javaPID = strings.TrimSpace(strings.Split(string(output), "\n")[0])
	}

	if javaPID == "" {
		return fmt.Errorf("no Java process found in container")
	}

	// Validate PID is numeric
	if _, err := strconv.Atoi(javaPID); err != nil {
		return fmt.Errorf("invalid Java PID: %s", javaPID)
	}

	fmt.Printf("Found Java process with PID: %s\n", javaPID)
	fmt.Println("Running thread dump...")

	// Run jattach to get thread dump
	cmd = exec.Command("kubectl", "exec", podName, "-c", containerName, "--", "/tmp/jattach", javaPID, "threaddump")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("jattach command failed: %v", err)
	}

	return nil
}

func runHeapDump(podName, containerName string) error {
	fmt.Println("Finding Java process...")
	
	// Find Java PID using pidof first, then pgrep as fallback
	var javaPID string
	
	// Try pidof java
	cmd := exec.Command("kubectl", "exec", podName, "-c", containerName, "--", "pidof", "java")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		javaPID = strings.TrimSpace(string(output))
		// If multiple PIDs, take the first one
		if strings.Contains(javaPID, " ") {
			javaPID = strings.Split(javaPID, " ")[0]
		}
	} else {
		// Try pgrep as fallback
		cmd = exec.Command("kubectl", "exec", podName, "-c", containerName, "--", "pgrep", "-f", "java")
		output, err = cmd.Output()
		if err != nil || len(output) == 0 {
			return fmt.Errorf("no Java process found in container")
		}
		javaPID = strings.TrimSpace(strings.Split(string(output), "\n")[0])
	}

	if javaPID == "" {
		return fmt.Errorf("no Java process found in container")
	}

	// Validate PID is numeric
	if _, err := strconv.Atoi(javaPID); err != nil {
		return fmt.Errorf("invalid Java PID: %s", javaPID)
	}

	fmt.Printf("Found Java process with PID: %s\n", javaPID)
	fmt.Println("Creating heap dump...")

	// Generate timestamp for unique filename
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	remoteHeapDumpPath := fmt.Sprintf("/tmp/heapdump-%s.hprof", timestamp)
	
	// Run jattach to create heap dump
	cmd = exec.Command("kubectl", "exec", podName, "-c", containerName, "--", "/tmp/jattach", javaPID, "dumpheap", remoteHeapDumpPath)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("jattach dumpheap command failed: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("Heap dump created in container: %s\n", remoteHeapDumpPath)
	
	// Download heap dump file locally
	localHeapDumpPath := fmt.Sprintf("heapdump-%s-%s.hprof", podName, timestamp)
	fmt.Printf("Downloading heap dump to: %s\n", localHeapDumpPath)
	
	cmd = exec.Command("kubectl", "cp", fmt.Sprintf("%s:%s", podName, remoteHeapDumpPath), localHeapDumpPath, "-c", containerName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download heap dump: %v", err)
	}

	// Clean up heap dump file from container
	fmt.Println("Cleaning up heap dump from container...")
	cmd = exec.Command("kubectl", "exec", podName, "-c", containerName, "--", "rm", "-f", remoteHeapDumpPath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: failed to clean up heap dump from container: %v\n", err)
	}

	fmt.Printf("Heap dump successfully downloaded to: %s\n", localHeapDumpPath)
	return nil
}
