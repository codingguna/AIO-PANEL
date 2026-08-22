package ops

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// CronJob represents a scheduled task
type CronJob struct {
	ID       int    `json:"id"`
	Schedule string `json:"schedule"` // e.g. "0 2 * * *"
	Command  string `json:"command"`
	User     string `json:"user"`
	Comment  string `json:"comment"`
}

// ListCronJobs reads cron jobs for the current host in real-time
func ListCronJobs(ctx context.Context) ([]CronJob, error) {
	if runtime.GOOS != "linux" {
		return []CronJob{}, nil
	}

	cmd := exec.CommandContext(ctx, "crontab", "-l")
	out, err := cmd.Output()
	if err != nil {
		return []CronJob{}, nil // crontab might be empty
	}

	var jobs []CronJob
	scanner := bufio.NewScanner(bytes.NewReader(out))
	idx := 1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 6 {
			schedule := strings.Join(fields[0:5], " ")
			command := strings.Join(fields[5:], " ")
			jobs = append(jobs, CronJob{
				ID:       idx,
				Schedule: schedule,
				Command:  command,
				User:     "root",
			})
			idx++
		}
	}

	return jobs, nil
}

// CreateCronJob appends a new task to the user's crontab
func CreateCronJob(ctx context.Context, schedule, command string) error {
	if runtime.GOOS != "linux" {
		return nil
	}

	cmd := exec.CommandContext(ctx, "crontab", "-l")
	existing, _ := cmd.Output()

	newEntry := fmt.Sprintf("%s\n%s %s\n", strings.TrimSpace(string(existing)), schedule, command)
	setCmd := exec.CommandContext(ctx, "crontab", "-")
	setCmd.Stdin = strings.NewReader(strings.TrimSpace(newEntry) + "\n")
	return setCmd.Run()
}

// DeleteCronJob removes a cron job by its 1-based index
func DeleteCronJob(ctx context.Context, id int) error {
	if runtime.GOOS != "linux" {
		return nil
	}

	cmd := exec.CommandContext(ctx, "crontab", "-l")
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	var newLines []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	idx := 1

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) >= 6 {
			if idx != id {
				newLines = append(newLines, line)
			}
			idx++
		} else {
			newLines = append(newLines, line)
		}
	}

	setCmd := exec.CommandContext(ctx, "crontab", "-")
	setCmd.Stdin = strings.NewReader(strings.Join(newLines, "\n") + "\n")
	return setCmd.Run()
}
