package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// sessionDirName is where nvim session files are written, inside the project.
const sessionDirName = ".tmux-sessionizer.d"

// serverFromArgs returns the socket of a remote-UI client, i.e. the value of
// --server in `nvim --server SOCK --remote-ui` (how tv attaches its windows
// to one shared headless nvim).
func serverFromArgs(args []string) (string, bool) {
	for i, a := range args {
		switch {
		case a == "--server" && i+1 < len(args):
			return args[i+1], true
		case strings.HasPrefix(a, "--server="):
			return strings.TrimPrefix(a, "--server="), true
		}
	}
	return "", false
}

// runtimeDir is where nvim puts its default server sockets.
func runtimeDir() string {
	if d := os.Getenv("XDG_RUNTIME_DIR"); d != "" {
		return d
	}
	return fmt.Sprintf("/run/user/%d", os.Getuid())
}

// nvimServer finds the RPC socket of the nvim behind pid. A TUI nvim hands
// its server to an embedded child process, whose socket is named after the
// child's pid.
func (t *procTree) nvimServer(pid int) (string, bool) {
	if args := cmdlineOf(pid); len(args) > 0 {
		if sock, ok := serverFromArgs(args); ok {
			return sock, true
		}
	}
	for _, p := range append([]int{pid}, t.children[pid]...) {
		sock := filepath.Join(runtimeDir(), "nvim."+strconv.Itoa(p)+".0")
		if st, err := os.Stat(sock); err == nil && st.Mode()&os.ModeSocket != 0 {
			return sock, true
		}
	}
	return "", false
}

// saveNvimSession asks the nvim listening on server to write a session file.
// It gives up after a few seconds, as a busy nvim would never answer.
func saveNvimSession(server, file string) bool {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	expr := fmt.Sprintf("execute('mksession! ' . fnameescape('%s'))", strings.ReplaceAll(file, "'", "''"))
	if err := exec.CommandContext(ctx, "nvim", "--server", server, "--remote-expr", expr).Run(); err != nil {
		return false
	}
	st, err := os.Stat(file)
	return err == nil && st.Size() > 0
}

// isRemoteUI reports whether args describe a remote-UI client, as opposed to
// an nvim that merely uses --server for something else.
func isRemoteUI(args []string) bool {
	for _, a := range args {
		if a == "--remote-ui" {
			return true
		}
	}
	return false
}
