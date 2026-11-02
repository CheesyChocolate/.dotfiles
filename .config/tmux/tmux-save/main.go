// tmux-save writes the current tmux session's windows, panes, layouts and
// running programs (with lf state) to <project>/.tmux-sessionizer as a shell
// script of tmux commands.
//
// usage: tmux-save [project_dir]
//
//	project_dir defaults to the session's start dir (#{session_path}).
//
// Restore: in a session that has a single window, run
//
//	source <project>/.tmux-sessionizer
//
// (S and DIR default to the current session name and $PWD.)
//
// What each pane gets back on restore:
//   - a whitelisted program (see relaunch in proc.go) is typed and run
//     (lf also gets its directory, cursor, options, filter and selection;
//     nvim gets a session file written to <project>/.tmux-sessionizer.d/:
//     a classic nvim reopens with `nvim -S`, while clients of a shared
//     server (`--remote-ui`) get the server restarted on its socket with
//     the session, and reattach to it; claude resumes its conversation
//     with `claude --resume <id>` from the directory it was started in,
//     or starts fresh if it has no transcript yet);
//   - any other foreground command is typed but not run, so Enter is yours;
//   - an idle shell prompt gets nothing.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "tmux-save: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if os.Getenv("TMUX") == "" {
		return errors.New("not inside tmux")
	}
	if len(args) > 1 {
		return errors.New("usage: tmux-save [project_dir]")
	}

	session, err := tmuxOut("display-message", "-p", "#S")
	if err != nil {
		return fmt.Errorf("cannot read session name: %w", err)
	}
	// Trailing colon: a bare "=2" is not resolved as a session when the name
	// is numeric.
	target := "=" + session + ":"

	project := ""
	if len(args) == 1 {
		project = args[0]
	}
	if project == "" {
		project, err = tmuxOut("display-message", "-p", "-t", target, "#{session_path}")
		if err != nil {
			return fmt.Errorf("cannot read session path: %w", err)
		}
	}
	project, err = filepath.Abs(project)
	if err != nil {
		return err
	}
	if st, err := os.Stat(project); err != nil || !st.IsDir() {
		return fmt.Errorf("not a directory: %s", project)
	}

	g := &generator{project: project, target: target, tree: loadProcTree()}
	if err := g.generate(); err != nil {
		return err
	}

	out := filepath.Join(project, ".tmux-sessionizer")
	if err := os.WriteFile(out, []byte(g.out.String()), 0o644); err != nil {
		return err
	}
	fmt.Println("saved", out)
	return nil
}
