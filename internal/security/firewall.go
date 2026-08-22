package security

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// FirewallRule represents a numbered UFW rule
type FirewallRule struct {
	ID       int    `json:"id"`
	ToPort   string `json:"to_port"`
	Protocol string `json:"protocol"` // tcp, udp, any
	Action   string `json:"action"`   // ALLOW, DENY, REJECT
	FromIP   string `json:"from_ip"`  // Anywhere or specific CIDR
	Comment  string `json:"comment"`
}

// FirewallStatus represents the current UFW firewall state
type FirewallStatus struct {
	Active          bool           `json:"active"`
	DefaultIncoming string         `json:"default_incoming"` // deny, allow
	DefaultOutgoing string         `json:"default_outgoing"` // allow, deny
	Rules           []FirewallRule `json:"rules"`
}

var ufwRuleRegex = regexp.MustCompile(`^\[\s*(\d+)\]\s+([^\s]+)\s+(ALLOW|DENY|REJECT)\s+IN\s+(.*)$`)

// GetFirewallStatus queries UFW status and rules in real-time
func GetFirewallStatus(ctx context.Context) (*FirewallStatus, error) {
	status := &FirewallStatus{
		Active:          false,
		DefaultIncoming: "deny",
		DefaultOutgoing: "allow",
		Rules:           make([]FirewallRule, 0),
	}

	if runtime.GOOS != "linux" {
		return status, nil
	}

	if _, err := exec.LookPath("ufw"); err != nil {
		return status, nil // UFW not installed
	}

	cmd := exec.CommandContext(ctx, "ufw", "status", "numbered")
	out, err := cmd.Output()
	if err != nil {
		return status, nil
	}

	outputStr := string(out)
	status.Active = strings.Contains(outputStr, "Status: active")
	if !status.Active {
		return status, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		match := ufwRuleRegex.FindStringSubmatch(line)
		if len(match) >= 5 {
			id, _ := strconv.Atoi(match[1])
			target := match[2]
			action := match[3]
			from := match[4]

			proto := "any"
			port := target
			if strings.Contains(target, "/") {
				parts := strings.Split(target, "/")
				port = parts[0]
				proto = parts[1]
			}

			status.Rules = append(status.Rules, FirewallRule{
				ID:       id,
				ToPort:   port,
				Protocol: proto,
				Action:   action,
				FromIP:   from,
			})
		}
	}

	return status, nil
}

// AddFirewallRule adds a new port rule to UFW
func AddFirewallRule(ctx context.Context, port string, proto, action, fromIP, comment string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("UFW firewall is only available on Linux")
	}

	var cmdArgs []string
	cmdArgs = append(cmdArgs, strings.ToLower(action))

	if fromIP != "" && fromIP != "Anywhere" && fromIP != "0.0.0.0/0" {
		cmdArgs = append(cmdArgs, "from", fromIP, "to", "any", "port", port)
	} else {
		if proto != "" && proto != "any" {
			cmdArgs = append(cmdArgs, fmt.Sprintf("%s/%s", port, proto))
		} else {
			cmdArgs = append(cmdArgs, port)
		}
	}

	if comment != "" {
		cmdArgs = append(cmdArgs, "comment", comment)
	}

	cmd := exec.CommandContext(ctx, "ufw", cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ufw rule addition failed: %s (%w)", strings.TrimSpace(stderr.String()), err)
	}

	return nil
}

// DeleteFirewallRule deletes a UFW rule by numbered index
func DeleteFirewallRule(ctx context.Context, ruleID int) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("UFW firewall is only available on Linux")
	}

	cmd := exec.CommandContext(ctx, "ufw", "--force", "delete", strconv.Itoa(ruleID))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete ufw rule %d: %s (%w)", ruleID, strings.TrimSpace(stderr.String()), err)
	}

	return nil
}

// ToggleFirewall enables or disables UFW with anti-lockout SSH checks
func ToggleFirewall(ctx context.Context, enable bool, sshPort int) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("UFW firewall is only available on Linux")
	}

	if enable {
		// Anti-lockout safeguard: Ensure SSH port is allowed before enabling
		if sshPort <= 0 {
			sshPort = 22
		}
		_ = exec.CommandContext(ctx, "ufw", "allow", fmt.Sprintf("%d/tcp", sshPort), "comment", "AIO SSH Anti-Lockout").Run()

		cmd := exec.CommandContext(ctx, "ufw", "--force", "enable")
		return cmd.Run()
	}

	cmd := exec.CommandContext(ctx, "ufw", "disable")
	return cmd.Run()
}
