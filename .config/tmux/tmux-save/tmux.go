package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// tmuxRaw runs tmux and returns its stdout untouched.
func tmuxRaw(args ...string) (string, error) {
	out, err := exec.Command("tmux", args...).Output()
	return string(out), err
}

// tmuxOut runs tmux and returns stdout without its trailing newlines.
func tmuxOut(args ...string) (string, error) {
	out, err := tmuxRaw(args...)
	return strings.TrimRight(out, "\n"), err
}

type pane struct {
	index int
	path  string
	pid   int
}

type window struct {
	index  string
	name   string
	layout string
	active bool
}

func listWindows(target string) ([]window, error) {
	out, err := tmuxOut("list-windows", "-t", target, "-F",
		"#{window_index}\t#{window_name}\t#{window_layout}\t#{window_active}")
	if err != nil {
		return nil, fmt.Errorf("list-windows: %w", err)
	}
	var ws []window
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		f := strings.SplitN(line, "\t", 4)
		for len(f) < 4 {
			f = append(f, "")
		}
		ws = append(ws, window{index: f[0], name: f[1], layout: f[2], active: f[3] == "1"})
	}
	return ws, nil
}

// listPanes returns the panes of one window. A failure yields no panes, the
// window is then emitted empty.
func listPanes(target, winIndex string) []pane {
	out, err := tmuxOut("list-panes", "-t", target+winIndex, "-F",
		"#{pane_index}\t#{pane_current_path}\t#{pane_current_command}\t#{pane_pid}")
	if err != nil {
		return nil
	}
	var ps []pane
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		f := strings.SplitN(line, "\t", 4)
		for len(f) < 4 {
			f = append(f, "")
		}
		idx, _ := strconv.Atoi(f[0])
		pid, _ := strconv.Atoi(f[3])
		ps = append(ps, pane{index: idx, path: f[1], pid: pid})
	}
	return ps
}
