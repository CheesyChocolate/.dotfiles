package main

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

// relaunch maps a program name to the command typed into the restored pane.
// Anything else running is typed but not run; idle shells get nothing.
var relaunch = map[string]string{
	"nvim":   "nvim .",
	"lf":     "lf",
	"claude": "claude",
}

// parseStatPPID extracts the parent pid from the contents of /proc/PID/stat.
// The command name sits in parentheses and may itself contain spaces or
// parentheses, so fields are counted after the last ')'.
func parseStatPPID(stat string) (int, bool) {
	i := strings.LastIndex(stat, ")")
	if i < 0 {
		return 0, false
	}
	fields := strings.Fields(stat[i+1:])
	if len(fields) < 2 {
		return 0, false
	}
	ppid, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, false
	}
	return ppid, true
}

// procTree is a snapshot of parent to children relations from /proc.
type procTree struct {
	children map[int][]int
}

func loadProcTree() *procTree {
	t := &procTree{children: map[int][]int{}}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return t
	}
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		b, err := os.ReadFile("/proc/" + e.Name() + "/stat")
		if err != nil {
			continue
		}
		if ppid, ok := parseStatPPID(string(b)); ok {
			t.children[ppid] = append(t.children[ppid], pid)
		}
	}
	for _, kids := range t.children {
		sort.Ints(kids)
	}
	return t
}

func commOf(pid int) string {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/comm")
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(b), "\n")
}

// findProg returns the first allowlisted program in a process tree: the
// process itself, then its descendants depth-first, children in pid order.
// Needed because wrappers (e.g. lfub, a sh script that runs lf as a child)
// make tmux report the wrapper, not the program.
func (t *procTree) findProg(pid int) (name string, progPID int, ok bool) {
	comm := commOf(pid)
	if _, listed := relaunch[comm]; listed {
		return comm, pid, true
	}
	for _, child := range t.children[pid] {
		if name, progPID, ok := t.findProg(child); ok {
			return name, progPID, true
		}
	}
	return "", 0, false
}

// shells are processes that, when they are a pane's own foreground process,
// mean the pane is idle at a prompt.
var shells = map[string]bool{
	"sh": true, "bash": true, "zsh": true, "dash": true, "fish": true,
	"ksh": true, "ash": true, "csh": true, "tcsh": true,
}

// parseStatTPGID extracts the terminal's foreground process group id from
// the contents of /proc/PID/stat. Fields after the command name are: state,
// ppid, pgrp, session, tty_nr, tpgid.
func parseStatTPGID(stat string) (int, bool) {
	i := strings.LastIndex(stat, ")")
	if i < 0 {
		return 0, false
	}
	fields := strings.Fields(stat[i+1:])
	if len(fields) < 6 {
		return 0, false
	}
	tpgid, err := strconv.Atoi(fields[5])
	if err != nil {
		return 0, false
	}
	return tpgid, true
}

// foregroundCmd returns the command line of whatever is running in the
// foreground of a pane, or false if the pane is idle at a shell prompt.
func foregroundCmd(panePID int) ([]string, bool) {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(panePID) + "/stat")
	if err != nil {
		return nil, false
	}
	tpgid, ok := parseStatTPGID(string(b))
	if !ok || tpgid <= 0 {
		return nil, false
	}
	if tpgid == panePID && shells[commOf(panePID)] {
		return nil, false
	}
	args := cmdlineOf(tpgid)
	return args, len(args) > 0
}

// cmdlineOf returns the argument vector of a process, or nil.
func cmdlineOf(pid int) []string {
	raw, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil || len(raw) == 0 {
		return nil
	}
	return strings.Split(strings.TrimRight(string(raw), "\x00"), "\x00")
}
