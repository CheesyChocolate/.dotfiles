package main

import (
	"regexp"
	"strings"
)

// q single-quotes s for a POSIX shell.
func q(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// dq escapes s for use inside double quotes in the generated script.
func dq(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, `$`, `\$`)
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}

// lfrcEsc escapes s for an lfrc double-quoted string.
func lfrcEsc(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// hdEsc escapes s for an unquoted heredoc.
func hdEsc(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `$`, `\$`)
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}

// relToDir renders p for use inside an lfrc double-quoted string within an
// unquoted heredoc: anchored on $DIR when inside the project, so it expands
// at restore time.
func relToDir(project, p string) string {
	switch {
	case p == project:
		return "$DIR"
	case strings.HasPrefix(p, project+"/"):
		return "$DIR/" + hdEsc(lfrcEsc(strings.TrimPrefix(p, project+"/")))
	default:
		return hdEsc(lfrcEsc(p))
	}
}

// pathExpr renders p as a shell word, anchored on "$DIR" when inside the
// project.
func pathExpr(project, p string) string {
	switch {
	case p == project:
		return `"$DIR"`
	case strings.HasPrefix(p, project+"/"):
		return `"$DIR"/` + q(strings.TrimPrefix(p, project+"/"))
	default:
		return q(p)
	}
}

var plainWord = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./~-]+$`)

// shellWords joins args into one command line, quoting only what needs it.
func shellWords(args []string) string {
	words := make([]string, len(args))
	for i, a := range args {
		if plainWord.MatchString(a) {
			words[i] = a
		} else {
			words[i] = q(a)
		}
	}
	return strings.Join(words, " ")
}

// typedPath renders p for use inside a double-quoted argument of the
// generated script, where it is typed into a pane: "$DIR" is left to expand
// when the script runs, everything else is escaped.
func typedPath(project, p string) string {
	switch {
	case p == project:
		return `\"$DIR\"`
	case strings.HasPrefix(p, project+"/"):
		return `\"$DIR\"/` + dq(q(strings.TrimPrefix(p, project+"/")))
	default:
		return dq(q(p))
	}
}
