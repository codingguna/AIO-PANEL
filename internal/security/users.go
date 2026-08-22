package security

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// LinuxUser represents a user account on the Linux host
type LinuxUser struct {
	Username  string `json:"username"`
	UID       int    `json:"uid"`
	GID       int    `json:"gid"`
	HomeDir   string `json:"home_dir"`
	Shell     string `json:"shell"`
	IsSudo    bool   `json:"is_sudo"`
	HasSSHKey bool   `json:"has_ssh_key"`
	IsSystem  bool   `json:"is_system"`
}

// ListLinuxUsers reads system user accounts in read-only mode
func ListLinuxUsers(ctx context.Context) ([]LinuxUser, error) {
	if runtime.GOOS != "linux" {
		// On non-Linux (e.g. Windows during development), return current actual user
		u, err := user.Current()
		if err != nil {
			return []LinuxUser{}, nil
		}
		return []LinuxUser{
			{
				Username:  u.Username,
				UID:       1000,
				GID:       1000,
				HomeDir:   u.HomeDir,
				Shell:     "cmd.exe",
				IsSudo:    true,
				HasSSHKey: false,
				IsSystem:  false,
			},
		}, nil
	}

	sudoUsers := getSudoGroupMembers()

	f, err := os.Open("/etc/passwd")
	if err != nil {
		return []LinuxUser{}, err
	}
	defer f.Close()

	var users []LinuxUser
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}

		username := parts[0]
		uid, _ := strconv.Atoi(parts[2])
		gid, _ := strconv.Atoi(parts[3])
		home := parts[5]
		shell := parts[6]

		isSystem := uid < 1000 && uid != 0
		hasSSH := fileExists(filepath.Join(home, ".ssh", "authorized_keys"))
		isSudo := sudoUsers[username] || uid == 0

		users = append(users, LinuxUser{
			Username:  username,
			UID:       uid,
			GID:       gid,
			HomeDir:   home,
			Shell:     shell,
			IsSudo:    isSudo,
			HasSSHKey: hasSSH,
			IsSystem:  isSystem,
		})
	}

	return users, nil
}

func getSudoGroupMembers() map[string]bool {
	members := make(map[string]bool)
	f, err := os.Open("/etc/group")
	if err != nil {
		return members
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) >= 4 {
			grpName := parts[0]
			if grpName == "sudo" || grpName == "wheel" || grpName == "admin" {
				userList := strings.Split(parts[3], ",")
				for _, u := range userList {
					trimmed := strings.TrimSpace(u)
					if trimmed != "" {
						members[trimmed] = true
					}
				}
			}
		}
	}
	return members
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// CreateLinuxUser creates a new Linux system user (explicit admin action)
func CreateLinuxUser(ctx context.Context, username, shell string, isSudo bool) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("user management is only available on Linux")
	}

	if shell == "" {
		shell = "/bin/bash"
	}

	cmd := exec.CommandContext(ctx, "useradd", "-m", "-s", shell, username)
	if err := cmd.Run(); err != nil {
		return err
	}

	if isSudo {
		_ = exec.CommandContext(ctx, "usermod", "-aG", "sudo", username).Run()
	}

	return nil
}

// DeleteLinuxUser deletes a Linux user and their home directory
func DeleteLinuxUser(ctx context.Context, username string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("user management is only available on Linux")
	}

	cleanUser := strings.TrimSpace(username)
	if cleanUser == "" || cleanUser == "root" || cleanUser == "daemon" || cleanUser == "bin" {
		return fmt.Errorf("cannot delete protected system user: %s", username)
	}

	cmd := exec.CommandContext(ctx, "userdel", "-r", cleanUser)
	return cmd.Run()
}
