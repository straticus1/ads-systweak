package api

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ProcessEntry holds per-process metadata returned by /api/processes.
type ProcessEntry struct {
	PID     int     `json:"pid"`
	PPID    int     `json:"ppid"`
	User    string  `json:"user"`
	CPU     float64 `json:"cpu"`
	Mem     float64 `json:"mem"`
	VSZ     int64   `json:"vsz"`
	RSS     int64   `json:"rss"`
	State   string  `json:"state"`
	Started string  `json:"started"`
	Time    string  `json:"time"`
	Command string  `json:"command"`
	Name    string  `json:"name"`
}

// ProcessDetail extends ProcessEntry with open-file information.
type ProcessDetail struct {
	ProcessEntry
	OpenFiles []string `json:"openFiles"`
	Dylibs    []string `json:"dylibs"`
	Sockets   []string `json:"sockets"`
	Pipes     []string `json:"pipes"`
	EnvVars   []string `json:"envVars"`
}

// HandleGetProcesses serves GET /api/processes.
func HandleGetProcesses(w http.ResponseWriter, r *http.Request) {
	// First pass: metadata (all single-token fields).
	metaOut, err := exec.Command("ps", "-axo",
		"pid=,ppid=,user=,%cpu=,%mem=,vsz=,rss=,state=,start=,time=,comm=").Output()
	if err != nil {
		http.Error(w, "ps failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Second pass: full command lines.
	cmdOut, _ := exec.Command("ps", "-axo", "pid=,command=").Output()
	cmdMap := parseCmdLines(string(cmdOut))

	entries := parseProcessList(string(metaOut), cmdMap)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PID < entries[j].PID
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// HandleGetProcessDetail serves GET /api/processes/{pid}.
func HandleGetProcessDetail(w http.ResponseWriter, r *http.Request) {
	pidStr := r.PathValue("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil || pid <= 0 {
		http.Error(w, "invalid pid", http.StatusBadRequest)
		return
	}

	// Fetch single-process metadata.
	metaOut, err := exec.Command("ps", "-p", pidStr, "-o",
		"pid=,ppid=,user=,%cpu=,%mem=,vsz=,rss=,state=,start=,time=,comm=").Output()
	if err != nil {
		http.Error(w, "process not found", http.StatusNotFound)
		return
	}

	cmdOut, _ := exec.Command("ps", "-p", pidStr, "-o", "pid=,command=").Output()
	cmdMap := parseCmdLines(string(cmdOut))

	entries := parseProcessList(string(metaOut), cmdMap)
	if len(entries) == 0 {
		http.Error(w, "process not found", http.StatusNotFound)
		return
	}
	entry := entries[0]

	detail := ProcessDetail{ProcessEntry: entry}
	detail.OpenFiles, detail.Dylibs, detail.Sockets, detail.Pipes = parseLsof(pidStr)
	detail.EnvVars = parseEnvVars(pidStr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

// parseCmdLines builds a map of PID → full command line from `ps -axo pid=,command=`.
func parseCmdLines(output string) map[int]string {
	m := make(map[int]string)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}
		m[pid] = strings.TrimSpace(parts[1])
	}
	return m
}

// parseProcessList converts ps output lines into ProcessEntry slice.
// Expected columns: pid ppid user %cpu %mem vsz rss state start time comm
func parseProcessList(output string, cmdMap map[int]string) []ProcessEntry {
	var entries []ProcessEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		// Minimum: pid ppid user cpu mem vsz rss state start time comm = 11 fields
		if len(fields) < 11 {
			continue
		}

		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		ppid, _ := strconv.Atoi(fields[1])
		user := fields[2]
		cpu, _ := strconv.ParseFloat(fields[3], 64)
		mem, _ := strconv.ParseFloat(fields[4], 64)
		vsz, _ := strconv.ParseInt(fields[5], 10, 64)
		rss, _ := strconv.ParseInt(fields[6], 10, 64)
		state := fields[7]
		started := fields[8]
		time := fields[9]
		// comm is the last field; if there are more fields (unlikely with comm=), join them.
		name := strings.Join(fields[10:], " ")

		cmd := cmdMap[pid]

		entries = append(entries, ProcessEntry{
			PID:     pid,
			PPID:    ppid,
			User:    user,
			CPU:     cpu,
			Mem:     mem,
			VSZ:     vsz,
			RSS:     rss,
			State:   state,
			Started: started,
			Time:    time,
			Command: cmd,
			Name:    name,
		})
	}
	return entries
}

// parseLsof runs lsof and categorises file descriptors for the given pid.
func parseLsof(pidStr string) (openFiles, dylibs, sockets, pipes []string) {
	out, err := exec.Command("lsof", "-p", pidStr, "-n", "-P").Output()
	if err != nil {
		return
	}

	dyLibSet := make(map[string]struct{})

	lines := strings.Split(string(out), "\n")
	// Skip header line.
	if len(lines) > 0 {
		lines = lines[1:]
	}

	const cap = 500

	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		// lsof columns: COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
		// Minimum 9 fields; NAME may contain spaces.
		if len(fields) < 9 {
			continue
		}
		fdType := fields[4]
		name := strings.Join(fields[8:], " ")

		switch {
		case fdType == "REG":
			if len(openFiles) < cap {
				openFiles = append(openFiles, name)
			}
		case fdType == "DYLIB" || strings.HasSuffix(name, ".dylib"):
			if _, seen := dyLibSet[name]; !seen {
				dyLibSet[name] = struct{}{}
				if len(dylibs) < cap {
					dylibs = append(dylibs, name)
				}
			}
		case fdType == "IPv4" || fdType == "IPv6":
			if len(sockets) < cap {
				sockets = append(sockets, name)
			}
		case fdType == "PIPE":
			if len(pipes) < cap {
				pipes = append(pipes, name)
			}
		}
	}
	return
}

var envVarRe = regexp.MustCompile(`[A-Z_][A-Z0-9_]*=\S+`)

// parseEnvVars attempts to extract KEY=VALUE pairs from ps ewww output.
func parseEnvVars(pidStr string) []string {
	out, err := exec.Command("ps", "ewww", "-p", pidStr).Output()
	if err != nil {
		return nil
	}
	matches := envVarRe.FindAllString(string(out), -1)
	if matches == nil {
		return []string{}
	}
	return matches
}
