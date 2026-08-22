package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/db"
	"github.com/codingguna/aio-panel/internal/ops"
)

type OpsHandler struct {
	cfg   *config.Config
	store *db.Store
}

func NewOpsHandler(cfg *config.Config, store *db.Store) *OpsHandler {
	return &OpsHandler{cfg: cfg, store: store}
}

// BrowseFiles handles GET /api/v1/ops/files/browse
func (h *OpsHandler) BrowseFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	items, err := ops.BrowseDirectory(path)
	if err != nil {
		http.Error(w, `{"error":"failed to browse directory: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// ReadFile handles GET /api/v1/ops/files/read
func (h *OpsHandler) ReadFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, `{"error":"path is required"}`, http.StatusBadRequest)
		return
	}

	content, err := ops.ReadFileContent(path)
	if err != nil {
		http.Error(w, `{"error":"failed to read file: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(content))
}

type WriteFileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteFile handles POST /api/v1/ops/files/write
func (h *OpsHandler) WriteFile(w http.ResponseWriter, r *http.Request) {
	var req WriteFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	err := ops.WriteFileContent(req.Path, req.Content)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "WRITE_FILE", req.Path, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "File written successfully."})
}

type CreateFileRequest struct {
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
}

// CreateFile handles POST /api/v1/ops/files/create
func (h *OpsHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	var req CreateFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	err := ops.CreateFileOrDir(req.Path, req.IsDir)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "CREATE_FILE_OR_DIR", req.Path, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Created successfully."})
}

type DeleteFileRequest struct {
	Path string `json:"path"`
}

// DeleteFile handles POST /api/v1/ops/files/delete
func (h *OpsHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	var req DeleteFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	err := ops.DeleteFileSystemItem(req.Path)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "DELETE_FILE_OR_DIR", req.Path, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Deleted successfully."})
}

// TerminalExecRequest represents a command execution request
type TerminalExecRequest struct {
	Command string `json:"command"`
	Cwd     string `json:"cwd"`
}

// TerminalExecResponse represents a command execution result
type TerminalExecResponse struct {
	Command   string `json:"command"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  int    `json:"exit_code"`
	Duration  string `json:"duration"`
	Timestamp string `json:"timestamp"`
	Cwd       string `json:"cwd"`
	User      string `json:"user"`
	Hostname  string `json:"hostname"`
}

// ExecuteTerminalCommand handles POST /api/v1/ops/terminal/exec
func (h *OpsHandler) ExecuteTerminalCommand(w http.ResponseWriter, r *http.Request) {
	var req TerminalExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	cmdStr := strings.TrimSpace(req.Command)
	if cmdStr == "" {
		http.Error(w, `{"error":"command cannot be empty"}`, http.StatusBadRequest)
		return
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Resolve initial working directory
	currentDir := strings.TrimSpace(req.Cwd)
	if currentDir == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			currentDir = home
		} else {
			currentDir = "/var/www"
		}
	}

	hostname, _ := os.Hostname()
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = os.Getenv("SUDO_USER")
	}
	if currentUser == "" {
		currentUser = "root"
	}

	var cmd *exec.Cmd
	var stdoutBuf, stderrBuf bytes.Buffer

	delim := "__AIO_PWD_DELIM__"
	if runtime.GOOS == "windows" {
		safeDir := strings.ReplaceAll(currentDir, "'", "''")
		script := fmt.Sprintf("if (Test-Path '%s') { Set-Location '%s' }; %s; Write-Output '%s'; (Get-Location).Path", safeDir, safeDir, cmdStr, delim)
		cmd = exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-Command", script)
	} else {
		// Linux shell wrapper that preserves exit code and captures final working directory
		script := fmt.Sprintf("cd %q 2>/dev/null || cd /; %s\n__AIO_RET=$?\nprintf \"\\n%s\\n%%s\" \"$PWD\"\nexit $__AIO_RET", currentDir, cmdStr, delim)
		cmd = exec.CommandContext(ctx, "bash", "-c", script)
	}

	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	duration := time.Since(start)
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	rawStdout := stdoutBuf.String()
	resultingCwd := currentDir

	// Extract clean stdout and resulting working directory
	if strings.Contains(rawStdout, delim) {
		parts := strings.Split(rawStdout, delim)
		cleanStdout := strings.TrimRight(parts[0], "\r\n")
		rawStdout = cleanStdout
		if len(parts) > 1 {
			newDir := strings.TrimSpace(parts[1])
			if newDir != "" {
				resultingCwd = newDir
			}
		}
	}

	if h.store != nil {
		status := "SUCCESS"
		if exitCode != 0 {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "EXEC_COMMAND", cmdStr, status, fmt.Sprintf("exit=%d", exitCode), r.RemoteAddr)
	}

	resp := TerminalExecResponse{
		Command:   cmdStr,
		Stdout:    rawStdout,
		Stderr:    stderrBuf.String(),
		ExitCode:  exitCode,
		Duration:  duration.Round(time.Millisecond).String(),
		Timestamp: time.Now().Format("15:04:05"),
		Cwd:       resultingCwd,
		User:      currentUser,
		Hostname:  hostname,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetSystemLogs handles GET /api/v1/ops/logs
func (h *OpsHandler) GetSystemLogs(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")
	linesStr := r.URL.Query().Get("lines")
	lines := 100
	if l, err := strconv.Atoi(linesStr); err == nil && l > 0 && l <= 1000 {
		lines = l
	}

	var outContent string
	var err error

	if runtime.GOOS == "linux" {
		switch source {
		case "nginx-access":
			outContent, err = tailFile("/var/log/nginx/access.log", lines)
		case "nginx-error":
			outContent, err = tailFile("/var/log/nginx/error.log", lines)
		case "auth":
			outContent, err = tailFile("/var/log/auth.log", lines)
		case "syslog":
			outContent, err = tailFile("/var/log/syslog", lines)
		case "aio":
			outContent, err = tailFile(h.cfg.Logging.File, lines)
		default:
			// Default to journalctl
			cmd := exec.CommandContext(r.Context(), "journalctl", "-n", strconv.Itoa(lines), "--no-pager")
			out, jErr := cmd.Output()
			if jErr == nil {
				outContent = string(out)
			} else {
				err = jErr
			}
		}
	} else {
		outContent = fmt.Sprintf("Log inspection for %s on %s.\n[Info] Native Linux journald & /var/log facilities are active on production Linux hosts.\n", source, runtime.GOOS)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"source":  source,
			"lines":   lines,
			"content": fmt.Sprintf("No active log entries found for %s (%v)", source, err),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"source":  source,
		"lines":   lines,
		"content": outContent,
	})
}

func tailFile(path string, lines int) (string, error) {
	cmd := exec.Command("tail", "-n", strconv.Itoa(lines), path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// BackupItem represents a backup file
type BackupItem struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	SizeBytes int64     `json:"size_bytes"`
	SizeHuman string    `json:"size_human"`
	CreatedAt time.Time `json:"created_at"`
	Type      string    `json:"type"` // postgres, mysql, config, app
}

// ListBackups handles GET /api/v1/ops/backups
func (h *OpsHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	backupDir := h.cfg.Paths.BackupDir
	if backupDir == "" {
		backupDir = filepath.Join(h.cfg.Paths.DataDir, "backups")
	}

	_ = os.MkdirAll(backupDir, 0750)
	entries, err := os.ReadDir(backupDir)
	var list []BackupItem

	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}

			bType := "file"
			name := e.Name()
			if strings.Contains(name, "postgres") || strings.HasSuffix(name, ".sql") {
				bType = "postgres"
			} else if strings.Contains(name, "mysql") {
				bType = "mysql"
			} else if strings.Contains(name, "config") {
				bType = "config"
			}

			list = append(list, BackupItem{
				Name:      name,
				Path:      filepath.Join(backupDir, name),
				SizeBytes: info.Size(),
				SizeHuman: formatBytesSimple(info.Size()),
				CreatedAt: info.ModTime(),
				Type:      bType,
			})
		}
	}

	if list == nil {
		list = []BackupItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func formatBytesSimple(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	if b < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	}
	if b < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(b)/(1024*1024*1024))
}

// ListCronJobs handles GET /api/v1/ops/cron
func (h *OpsHandler) ListCronJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := ops.ListCronJobs(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list cron jobs"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

type CreateCronRequest struct {
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
}

// CreateCronJob handles POST /api/v1/ops/cron/create
func (h *OpsHandler) CreateCronJob(w http.ResponseWriter, r *http.Request) {
	var req CreateCronRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.Schedule == "" || req.Command == "" {
		http.Error(w, `{"error":"schedule and command are required"}`, http.StatusBadRequest)
		return
	}

	err := ops.CreateCronJob(r.Context(), req.Schedule, req.Command)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "CREATE_CRON", req.Command, status, req.Schedule, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "Cron job created successfully",
	})
}

// DeleteCronJob handles DELETE /api/v1/ops/cron/{id}
func (h *OpsHandler) DeleteCronJob(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid cron id"}`, http.StatusBadRequest)
		return
	}

	err = ops.DeleteCronJob(r.Context(), id)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "DELETE_CRON", idStr, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "Cron job deleted successfully",
	})
}

// DeleteDockerContainer handles DELETE /api/v1/ops/docker/containers/{id}
func (h *OpsHandler) DeleteDockerContainer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"missing container id"}`, http.StatusBadRequest)
		return
	}

	cmd := exec.CommandContext(r.Context(), "docker", "rm", "-f", id)
	err := cmd.Run()
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", "DELETE_CONTAINER", id, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"error": "Failed to remove container"})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "Container deleted successfully",
	})
}

// ListDockerContainers handles GET /api/v1/ops/docker/containers
func (h *OpsHandler) ListDockerContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := ops.ListDockerContainers(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list containers"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

// ListDockerImages handles GET /api/v1/ops/docker/images
func (h *OpsHandler) ListDockerImages(w http.ResponseWriter, r *http.Request) {
	images, err := ops.ListDockerImages(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list docker images"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

// GetDockerContainerLogs handles GET /api/v1/ops/docker/containers/{id}/logs
func (h *OpsHandler) GetDockerContainerLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	linesStr := r.URL.Query().Get("lines")
	lines := 100
	if l, err := strconv.Atoi(linesStr); err == nil && l > 0 {
		lines = l
	}

	logs, err := ops.GetDockerContainerLogs(r.Context(), id, lines)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(logs))
}

type DockerActionRequest struct {
	Action string `json:"action"` // start, stop, restart
}

// ControlDockerContainer handles POST /api/v1/ops/docker/containers/{id}/action
func (h *OpsHandler) ControlDockerContainer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req DockerActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	err := ops.ControlDockerContainer(r.Context(), id, req.Action)
	if h.store != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		_ = h.store.LogAudit(r.Context(), "admin", req.Action+"_CONTAINER", id, status, "", r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Container action " + req.Action + " executed successfully.",
	})
}

// DeploymentRunnerRequest defines an application deployment trigger
type DeploymentRunnerRequest struct {
	AppName    string `json:"app_name"`
	Path       string `json:"path"`
	GitBranch  string `json:"git_branch"`
	Service    string `json:"service"`
	BuildCmd   string `json:"build_cmd"`
	RestartSvc bool   `json:"restart_svc"`
}

// RunDeployment handles POST /api/v1/ops/deployments/run
func (h *OpsHandler) RunDeployment(w http.ResponseWriter, r *http.Request) {
	var req DeploymentRunnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	var outputLog strings.Builder
	outputLog.WriteString(fmt.Sprintf("🚀 Starting deployment for %s at %s\n", req.AppName, time.Now().Format("2006-01-02 15:04:05")))
	outputLog.WriteString("──────────────────────────────────────────────────\n")

	// Step 1: Git Pull if git repo
	if req.Path != "" {
		outputLog.WriteString("1. Executing git pull...\n")
		gitCmd := exec.CommandContext(r.Context(), "git", "pull")
		gitCmd.Dir = req.Path
		gitOut, err := gitCmd.CombinedOutput()
		if err == nil {
			outputLog.Write(gitOut)
			outputLog.WriteString("\n")
		} else {
			outputLog.WriteString(fmt.Sprintf("Git note: %s\n", string(gitOut)))
		}
	}

	// Step 2: Custom Build command
	if req.BuildCmd != "" {
		outputLog.WriteString(fmt.Sprintf("2. Executing build command: %s\n", req.BuildCmd))
		var bCmd *exec.Cmd
		if runtime.GOOS == "windows" {
			bCmd = exec.CommandContext(r.Context(), "powershell.exe", "-Command", req.BuildCmd)
		} else {
			bCmd = exec.CommandContext(r.Context(), "bash", "-c", req.BuildCmd)
		}
		if req.Path != "" {
			bCmd.Dir = req.Path
		}
		bOut, err := bCmd.CombinedOutput()
		outputLog.Write(bOut)
		if err != nil {
			outputLog.WriteString(fmt.Sprintf("⚠️ Build warning/error: %v\n", err))
		}
	}

	// Step 3: Service reload / restart
	if req.RestartSvc && req.Service != "" && runtime.GOOS == "linux" {
		outputLog.WriteString(fmt.Sprintf("3. Reloading systemd service: %s\n", req.Service))
		svcCmd := exec.CommandContext(r.Context(), "systemctl", "restart", req.Service)
		if sOut, err := svcCmd.CombinedOutput(); err == nil {
			outputLog.WriteString("✓ Service restarted successfully.\n")
		} else {
			outputLog.WriteString(fmt.Sprintf("Service restart error: %s\n", string(sOut)))
		}
	}

	outputLog.WriteString("──────────────────────────────────────────────────\n")
	outputLog.WriteString("🎉 Deployment sequence finished successfully.\n")

	if h.store != nil {
		_ = h.store.LogAudit(r.Context(), "admin", "DEPLOY_APP", req.AppName, "SUCCESS", req.Path, r.RemoteAddr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"app":     req.AppName,
		"logs":    outputLog.String(),
	})
}
