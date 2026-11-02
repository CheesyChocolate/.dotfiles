package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestQuote(t *testing.T) {
	cases := []struct {
		fn       string
		in, want string
	}{
		{"q", "plain", `'plain'`},
		{"q", "it's", `'it'\''s'`},
		{"dq", `a"b$c` + "`d" + `\e`, `a\"b\$c\` + "`d" + `\\e`},
		{"lfrc", `a"b\c`, `a\"b\\c`},
		{"hd", "a$b`c" + `\d"`, `a\$b\` + "`c" + `\\d"`},
	}
	for _, c := range cases {
		var got string
		switch c.fn {
		case "q":
			got = q(c.in)
		case "dq":
			got = dq(c.in)
		case "lfrc":
			got = lfrcEsc(c.in)
		case "hd":
			got = hdEsc(c.in)
		}
		if got != c.want {
			t.Errorf("%s(%q) = %q, want %q", c.fn, c.in, got, c.want)
		}
	}
}

func TestRelToDir(t *testing.T) {
	const proj = "/home/u/proj"
	cases := []struct{ in, want string }{
		{proj, "$DIR"},
		{proj + "/sub dir/it's", "$DIR/sub dir/it's"},
		{proj + `/a"b`, `$DIR/a\\"b`}, // lfrc gives a\"b, heredoc doubles the backslash
		{proj + "/a$b", `$DIR/a\$b`},
		{"/elsewhere/x", "/elsewhere/x"},
		{"/home/u/proj-other/x", "/home/u/proj-other/x"}, // prefix match must respect "/"
	}
	for _, c := range cases {
		if got := relToDir(proj, c.in); got != c.want {
			t.Errorf("relToDir(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPathExpr(t *testing.T) {
	const proj = "/home/u/proj"
	cases := []struct{ in, want string }{
		{proj, `"$DIR"`},
		{proj + "/it's", `"$DIR"/'it'\''s'`},
		{"/elsewhere", `'/elsewhere'`},
		{"/home/u/proj-other", `'/home/u/proj-other'`},
	}
	for _, c := range cases {
		if got := pathExpr(proj, c.in); got != c.want {
			t.Errorf("pathExpr(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFilterFromLine(t *testing.T) {
	const info = "-rw-r--r-- 1 u u 0B Sat Oct 31 15:40:46 2026"
	cases := []struct {
		name, line, file, want string
		ok                     bool
	}{
		{"filter", info + "                    txt   2/4", "/p/b.txt", "txt", true},
		{"filter and selection", info + "          2    txt   3/6", "/p/b.txt", "txt", true},
		{"no filter", info + "                         3/6", "/p/b.txt", "", false},
		{"selection only", info + "                 2   3/6", "/p/b.txt", "", false},
		{"numeric filter is ambiguous", info + "      2024   1/3", "/p/2024.txt", "", false},
		{"cursor file does not match", info + "       zzz   1/3", "/p/b.txt", "", false},
		{"case insensitive", info + "                 TXT   1/3", "/p/b.txt", "TXT", true},
		{"semicolon rejected", info + "              a;b   1/3", "/p/a;b.txt", "", false},
		{"filter with single space", info + "     foo bar   1/3", "/p/foo bar.txt", "foo bar", true},
		{"message instead of ruler", "error: something went wrong", "/p/b.txt", "", false},
		{"empty line", "", "/p/b.txt", "", false},
	}
	for _, c := range cases {
		got, ok := filterFromLine(c.line, c.file)
		if got != c.want || ok != c.ok {
			t.Errorf("%s: got (%q, %v), want (%q, %v)", c.name, got, ok, c.want, c.ok)
		}
	}
}

func TestLastLine(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a\nb\n", "b"},
		{"a\nb", "b"},
		{"a\n\n", ""}, // like tail -n 1: an empty last line counts
		{"", ""},
		{"only\n", "only"},
	}
	for _, c := range cases {
		if got := lastLine(c.in); got != c.want {
			t.Errorf("lastLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseLfState(t *testing.T) {
	got := parseLfState("/p/d\n/p/d/b.txt\ntrue\nsize\ntrue\nfalse\n/p/d/c d.txt\n/p/d/it's.txt\n")
	want := &lfState{
		cwd: "/p/d", file: "/p/d/b.txt", hidden: "true", sortby: "size",
		reverse: "true", dirfirst: "false",
		selected: []string{"/p/d/c d.txt", "/p/d/it's.txt"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	// Empty selection and empty cursor file leave no phantom entries.
	got = parseLfState("/p/d\n\nfalse\nnatural\nfalse\ntrue\n\n")
	if got.file != "" || len(got.selected) != 0 || got.sortby != "natural" {
		t.Errorf("empty case parsed wrong: %+v", got)
	}
}

func TestParseStatPPID(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"123 (zsh) S 45 123 123 0 -1", 45, true},
		{"123 (my (odd) name) S 77 1 1 0", 77, true}, // parens and spaces in comm
		{"123 (x y) S", 0, false},
		{"garbage", 0, false},
	}
	for _, c := range cases {
		got, ok := parseStatPPID(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("parseStatPPID(%q) = (%d, %v), want (%d, %v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestShellWords(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{[]string{"/bin/sh", "./run.sh"}, "/bin/sh ./run.sh"},
		{[]string{"python3", "my script.py", "--n=3"}, "python3 'my script.py' --n=3"},
		{[]string{"echo", "it's", "$HOME", "a;b"}, `echo 'it'\''s' '$HOME' 'a;b'`},
		{[]string{"node", ""}, "node ''"},
	}
	for _, c := range cases {
		if got := shellWords(c.in); got != c.want {
			t.Errorf("shellWords(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseStatTPGID(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"10 (zsh) S 9 10 10 34816 4242 4194304", 4242, true},
		{"10 (a (b) c) S 9 10 10 34816 -1 0", -1, true},
		{"10 (zsh) S 9 10", 0, false},
		{"garbage", 0, false},
	}
	for _, c := range cases {
		got, ok := parseStatTPGID(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("parseStatTPGID(%q) = (%d, %v), want (%d, %v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestServerFromArgs(t *testing.T) {
	cases := []struct {
		in   []string
		want string
		ok   bool
	}{
		{[]string{"nvim", "--server", "/run/x.sock", "--remote-ui"}, "/run/x.sock", true},
		{[]string{"nvim", "--server=/run/y.sock", "--remote-ui"}, "/run/y.sock", true},
		{[]string{"nvim", "a.txt"}, "", false},
		{[]string{"nvim", "--server"}, "", false},
	}
	for _, c := range cases {
		got, ok := serverFromArgs(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("serverFromArgs(%q) = (%q, %v), want (%q, %v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestIsRemoteUI(t *testing.T) {
	if !isRemoteUI([]string{"nvim", "--server", "/s", "--remote-ui"}) {
		t.Error("remote-ui client not detected")
	}
	if isRemoteUI([]string{"nvim", "--server", "/s", "--remote-expr", "1"}) {
		t.Error("--remote-expr misdetected as a UI client")
	}
}

func TestSharedVar(t *testing.T) {
	g := &generator{}
	v1, first1 := g.sharedVar("/a.sock")
	v1b, first1b := g.sharedVar("/a.sock")
	v2, first2 := g.sharedVar("/b.sock")
	if v1 != "_nvsock1" || !first1 || v1b != v1 || first1b || v2 != "_nvsock2" || !first2 {
		t.Errorf("got (%s,%v) (%s,%v) (%s,%v)", v1, first1, v1b, first1b, v2, first2)
	}
}

func TestTypedPath(t *testing.T) {
	const proj = "/home/u/proj"
	cases := []struct{ in, want string }{
		{proj, `\"$DIR\"`},
		{proj + "/it's", `\"$DIR\"/'it'\\''s'`},
		{"/elsewhere/x y", `'/elsewhere/x y'`},
		{"/home/u/proj-other", `'/home/u/proj-other'`},
	}
	for _, c := range cases {
		if got := typedPath(proj, c.in); got != c.want {
			t.Errorf("typedPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSessionIDFromLink(t *testing.T) {
	id := "7aba38b6-ef5a-42d9-be74-d43f45b16730"
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"/tmp/claude-1000/-home-suren/" + id + "/tasks", id, true},
		{"/tmp/claude-1000/-home-suren--config-yay/" + id + "/tasks", id, true},
		{"/tmp/other/" + id + "/tasks", "", false},
		{"/tmp/claude-1000/-home-suren/" + id + "/other", "", false},
	}
	for _, c := range cases {
		got, ok := sessionIDFromLink(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("sessionIDFromLink(%q) = (%q, %v)", c.in, got, ok)
		}
	}
}

func TestEncodeProjectDir(t *testing.T) {
	cases := map[string]string{
		"/home/suren":                  "-home-suren",
		"/home/suren/.config/yay":      "-home-suren--config-yay",
		"/home/suren/Templates/AI/x y": "-home-suren-Templates-AI-x-y",
	}
	for in, want := range cases {
		if got := encodeProjectDir(in); got != want {
			t.Errorf("encodeProjectDir(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStartDirFromTranscript(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "-home-suren")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "abc.jsonl")
	// First record has a different cwd, as after a `cd`; the right one follows.
	body := `{"type":"summary"}` + "\n" +
		`{"cwd":"/home/suren/Trash","type":"user"}` + "\n" +
		`{"cwd":"/home/suren","type":"user"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := startDirFromTranscript(path)
	if !ok || got != "/home/suren" {
		t.Errorf("got (%q, %v), want /home/suren", got, ok)
	}
	if _, ok := startDirFromTranscript(filepath.Join(dir, "missing.jsonl")); ok {
		t.Error("missing transcript reported ok")
	}
}
