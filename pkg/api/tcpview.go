package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// SocketEntry represents a single network socket.
type SocketEntry struct {
	PID         int    `json:"pid"`
	ProcessName string `json:"processName"`
	Protocol    string `json:"protocol"`   // TCP, UDP, TCP6, UDP6
	LocalAddr   string `json:"localAddr"`  // host:port
	RemoteAddr  string `json:"remoteAddr"` // host:port or "" if listening
	State       string `json:"state"`      // ESTABLISHED, LISTEN, CLOSE_WAIT, etc.
	User        string `json:"user"`       // username owning the socket
}

// HandleGetTCPView handles GET /api/tcpview.
func HandleGetTCPView(w http.ResponseWriter, r *http.Request) {
	seen := make(map[string]bool)
	var entries []SocketEntry

	for _, proto := range []string{"TCP", "UDP"} {
		parsed, err := runLsof(proto)
		if err != nil {
			// lsof unavailable or failed — skip this protocol
			continue
		}
		for _, e := range parsed {
			key := fmt.Sprintf("%d:%s:%s", e.PID, e.LocalAddr, e.RemoteAddr)
			if seen[key] {
				continue
			}
			seen[key] = true
			entries = append(entries, e)
		}
	}

	if entries == nil {
		entries = []SocketEntry{}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].PID != entries[j].PID {
			return entries[i].PID < entries[j].PID
		}
		return entries[i].LocalAddr < entries[j].LocalAddr
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// runLsof runs lsof for the given protocol ("TCP" or "UDP") and parses results.
func runLsof(proto string) ([]SocketEntry, error) {
	cmd := exec.Command("lsof", "-i", proto, "-n", "-P")
	out, err := cmd.Output()
	if err != nil {
		// lsof exits non-zero when it finds nothing; treat any exec error as empty
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, err
		}
	}

	var entries []SocketEntry
	lines := strings.Split(string(out), "\n")
	if len(lines) <= 1 {
		return entries, nil
	}

	// Skip header line
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split into at most 9 fields; the NAME column (9th) may contain spaces
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		// Columns: COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
		// Indices:     0    1    2   3    4      5        6    7    8+
		typeField := fields[4]
		if typeField != "IPv4" && typeField != "IPv6" {
			continue
		}

		// NAME is everything from index 8 onward (rejoin in case spaces exist)
		nameField := strings.Join(fields[8:], " ")

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		command := fields[0]
		user := fields[2]

		localAddr, remoteAddr, state := parseName(nameField)

		// Determine protocol label, incorporating IP version
		protocol := proto
		if typeField == "IPv6" {
			protocol = proto + "6"
		}

		entries = append(entries, SocketEntry{
			PID:         pid,
			ProcessName: command,
			Protocol:    protocol,
			LocalAddr:   localAddr,
			RemoteAddr:  remoteAddr,
			State:       state,
			User:        user,
		})
	}

	return entries, nil
}

// parseName parses an lsof NAME field such as:
//
//	"192.168.1.1:52341->93.184.216.34:443 (ESTABLISHED)"
//	"*:8080 (LISTEN)"
//	"[::1]:631 (LISTEN)"
func parseName(name string) (localAddr, remoteAddr, state string) {
	// Extract trailing (STATE) if present
	if idx := strings.LastIndex(name, " ("); idx != -1 {
		state = strings.TrimSuffix(name[idx+2:], ")")
		name = name[:idx]
	}

	if idx := strings.Index(name, "->"); idx != -1 {
		localAddr = name[:idx]
		remoteAddr = name[idx+2:]
	} else {
		localAddr = name
		remoteAddr = ""
	}

	return
}
