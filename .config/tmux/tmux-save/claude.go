package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// A running claude holds an open directory fd on
// /tmp/claude-<uid>/<encoded start dir>/<session id>/tasks, which is the
// only place its session id shows (it is not in argv or the environment).
var claudeTasksRe = regexp.MustCompile(
	`/claude-[0-9]+/.+/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})/tasks$`)

// sessionIDFromLink extracts the session id from an fd link target.
func sessionIDFromLink(link string) (string, bool) {
	m := claudeTasksRe.FindStringSubmatch(link)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// claudeSessionID finds the session id of a running claude process.
func claudeSessionID(pid int) (string, bool) {
	dir := "/proc/" + strconv.Itoa(pid) + "/fd"
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}
	for _, e := range entries {
		link, err := os.Readlink(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		if id, ok := sessionIDFromLink(link); ok {
			return id, true
		}
	}
	return "", false
}

// claudeConfigDir returns the config dir of a claude process: its own
// CLAUDE_CONFIG_DIR if set, else ours, else ~/.claude.
func claudeConfigDir(pid int) string {
	if b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/environ"); err == nil {
		for _, kv := range strings.Split(string(b), "\x00") {
			if v, ok := strings.CutPrefix(kv, "CLAUDE_CONFIG_DIR="); ok && v != "" {
				return v
			}
		}
	}
	if v := os.Getenv("CLAUDE_CONFIG_DIR"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

var nonAlnum = regexp.MustCompile(`[^A-Za-z0-9]`)

// encodeProjectDir is how claude names a project folder after the directory
// the session was started in.
func encodeProjectDir(cwd string) string {
	return nonAlnum.ReplaceAllString(cwd, "-")
}

// startDirFromTranscript recovers the directory a session was started in:
// the first "cwd" recorded in the transcript whose encoding matches the
// project folder the transcript lives in.
func startDirFromTranscript(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()
	want := filepath.Base(filepath.Dir(path))

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for i := 0; i < 500 && sc.Scan(); i++ {
		var rec struct {
			Cwd string `json:"cwd"`
		}
		if json.Unmarshal(sc.Bytes(), &rec) != nil || rec.Cwd == "" {
			continue
		}
		if encodeProjectDir(rec.Cwd) == want {
			return rec.Cwd, true
		}
	}
	return "", false
}

// claudeResume finds what is needed to resume the session of the claude
// process pid: its id and the directory to resume it from. It fails if the
// session has no transcript yet (nothing has been said, so nothing to
// resume).
func claudeResume(pid int) (id, startDir string, ok bool) {
	id, ok = claudeSessionID(pid)
	if !ok {
		return "", "", false
	}
	matches, _ := filepath.Glob(filepath.Join(claudeConfigDir(pid), "projects", "*", id+".jsonl"))
	for _, m := range matches {
		if dir, ok := startDirFromTranscript(m); ok {
			return id, dir, true
		}
	}
	return "", "", false
}
