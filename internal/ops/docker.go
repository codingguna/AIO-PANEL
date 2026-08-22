package ops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// DockerContainer represents a container instance
type DockerContainer struct {
	ID      string `json:"id"`
	Names   string `json:"names"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	State   string `json:"state"` // running, exited, paused
	Ports   string `json:"ports"`
	Created string `json:"created"`
}

// ListDockerContainers discovers running and stopped containers in real-time
func ListDockerContainers(ctx context.Context) ([]DockerContainer, error) {
	if runtime.GOOS != "linux" {
		// Check if docker executable is present on host
		if _, err := exec.LookPath("docker"); err != nil {
			return []DockerContainer{}, nil
		}
	}

	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--format", "{{json .}}")
	out, err := cmd.Output()
	if err != nil {
		return []DockerContainer{}, nil // Docker not active
	}

	var containers []DockerContainer
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		var raw struct {
			ID        string `json:"ID"`
			Names     string `json:"Names"`
			Image     string `json:"Image"`
			Status    string `json:"Status"`
			State     string `json:"State"`
			Ports     string `json:"Ports"`
			CreatedAt string `json:"CreatedAt"`
		}

		if err := json.Unmarshal([]byte(trimmed), &raw); err == nil {
			containers = append(containers, DockerContainer{
				ID:      raw.ID,
				Names:   raw.Names,
				Image:   raw.Image,
				Status:  raw.Status,
				State:   raw.State,
				Ports:   raw.Ports,
				Created: raw.CreatedAt,
			})
		}
	}

	return containers, nil
}

// ControlDockerContainer dispatches start, stop, restart to a container
func ControlDockerContainer(ctx context.Context, containerID, action string) error {
	cmd := exec.CommandContext(ctx, "docker", action, containerID)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker %s %s failed: %s (%w)", action, containerID, strings.TrimSpace(stderr.String()), err)
	}

	return nil
}

// DockerImage represents a Docker image on the host
type DockerImage struct {
	ID         string `json:"id"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Size       string `json:"size"`
	Created    string `json:"created"`
}

// ListDockerImages lists all local Docker images
func ListDockerImages(ctx context.Context) ([]DockerImage, error) {
	if _, err := exec.LookPath("docker"); err != nil {
		return []DockerImage{}, nil
	}

	cmd := exec.CommandContext(ctx, "docker", "images", "--format", "{{json .}}")
	out, err := cmd.Output()
	if err != nil {
		return []DockerImage{}, nil
	}

	var images []DockerImage
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		var raw struct {
			ID         string `json:"ID"`
			Repository string `json:"Repository"`
			Tag        string `json:"Tag"`
			Size       string `json:"Size"`
			CreatedAt  string `json:"CreatedAt"`
		}
		if err := json.Unmarshal([]byte(trimmed), &raw); err == nil {
			images = append(images, DockerImage{
				ID:         raw.ID,
				Repository: raw.Repository,
				Tag:        raw.Tag,
				Size:       raw.Size,
				Created:    raw.CreatedAt,
			})
		}
	}
	return images, nil
}

// GetDockerContainerLogs returns recent logs for a Docker container
func GetDockerContainerLogs(ctx context.Context, containerID string, lines int) (string, error) {
	if lines <= 0 || lines > 500 {
		lines = 100
	}
	cmd := exec.CommandContext(ctx, "docker", "logs", "--tail", fmt.Sprintf("%d", lines), containerID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to fetch container logs: %s (%w)", string(out), err)
	}
	return string(out), nil
}

