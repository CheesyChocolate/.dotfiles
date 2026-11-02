package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// lfState is what a live lf client reports about itself.
type lfState struct {
	cwd      string
	file     string
	hidden   string
	sortby   string
	reverse  string
	dirfirst string
	selected []string
}

// lfQueryFmt asks lf (via its remote interface) to print, one per line:
// directory, file under the cursor, hidden, sortby, reverse, dirfirst, then
// the selected files.
const lfQueryFmt = ` $printf "%s\n%s\n%s\n%s\n%s\n%s\n%s\n" "$PWD" "$f" "$lf_hidden" "$lf_sortby" "$lf_reverse" "$lf_dirfirst" "$fs" >`

// queryLf asks the lf client with the given pid (its client id) for its
// state. It fails if lf is missing or the client can't be reached in time.
func queryLf(pid int) (*lfState, bool) {
	if _, err := exec.LookPath("lf"); err != nil {
		return nil, false
	}
	f, err := os.CreateTemp("", "tmux-save-*")
	if err != nil {
		return nil, false
	}
	tmp := f.Name()
	f.Close()
	os.Remove(tmp)
	defer os.Remove(tmp)

	cmd := exec.Command("lf", "-remote", fmt.Sprintf("send %d", pid)+lfQueryFmt+tmp)
	if err := cmd.Run(); err != nil {
		return nil, false
	}
	for i := 0; i < 40; i++ {
		if st, err := os.Stat(tmp); err == nil && st.Size() > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	b, err := os.ReadFile(tmp)
	if err != nil || len(b) == 0 {
		return nil, false
	}
	return parseLfState(string(b)), true
}

// parseLfState parses the output of lfQueryFmt. Trailing newlines are
// dropped first, so an empty selection leaves no empty entry.
func parseLfState(raw string) *lfState {
	lines := strings.Split(strings.TrimRight(raw, "\n"), "\n")
	get := func(i int) string {
		if i < len(lines) {
			return lines[i]
		}
		return ""
	}
	st := &lfState{
		cwd: get(0), file: get(1), hidden: get(2),
		sortby: get(3), reverse: get(4), dirfirst: get(5),
	}
	if len(lines) > 6 {
		st.selected = lines[6:]
	}
	return st
}

var (
	spaceRun   = regexp.MustCompile(` {2,}`)
	posRe      = regexp.MustCompile(`^[0-9]+/[0-9]+$`)
	countRe    = regexp.MustCompile(`^[0-9]+$`)
	fileModeRe = regexp.MustCompile(`^[-dlcbps][-rwxsStT]{9}`)
)

// lastLine returns the last line of s the way `tail -n 1` would, even when
// that line is empty.
func lastLine(s string) string {
	s = strings.TrimSuffix(s, "\n")
	return s[strings.LastIndex(s, "\n")+1:]
}

// filterFromLine reads lf's active filter off the ruler line (bottom right,
// before the "n/m" position), since lf does not export it. It only trusts
// the text if the file under the cursor actually matches it.
func filterFromLine(line, file string) (string, bool) {
	var chunks []string
	for _, c := range spaceRun.Split(line, -1) {
		if c != "" {
			chunks = append(chunks, c)
		}
	}
	n := len(chunks)
	if n < 2 || !posRe.MatchString(chunks[n-1]) {
		return "", false
	}
	cand := chunks[n-2]
	if countRe.MatchString(cand) { // that's the selection count
		return "", false
	}
	if fileModeRe.MatchString(cand) { // that's the file info
		return "", false
	}
	if strings.Contains(cand, ";") {
		return "", false
	}
	base := file[strings.LastIndex(file, "/")+1:]
	if !strings.Contains(strings.ToLower(base), strings.ToLower(cand)) {
		return "", false
	}
	return cand, true
}

// lfFilterFromScreen reads the filter from the live pane's screen.
func lfFilterFromScreen(paneTarget, file string) (string, bool) {
	out, _ := tmuxRaw("capture-pane", "-p", "-t", paneTarget)
	return filterFromLine(lastLine(out), file)
}
