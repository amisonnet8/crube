// Package e2e checks the docs against what crube really does: the examples
// in docs/ (commands and their output, JSON, error lines) are read from the
// Markdown files and run, and the MCP server is checked to return what the
// CLI's --json gives.
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/amisonnet8/crube/internal/cli"
)

// block is a fenced code block of a Markdown file.
type block struct {
	lang    string // after the opening ``` ("json", or "")
	body    string // the lines inside, each ending in "\n"
	heading string // the last "## " heading above it
	index   int    // its position among the blocks under that heading
	before  string // the text between the previous block and this one
}

// readBlocks reads the code blocks of a file under docs/ (or the root).
func readBlocks(t *testing.T, name string) []block {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	var (
		blocks  []block
		cur     *block
		heading string
		count   = map[string]int{}
		text    strings.Builder
	)
	for _, line := range strings.Split(string(data), "\n") {
		if cur != nil {
			if strings.HasPrefix(line, "```") {
				blocks = append(blocks, *cur)
				cur = nil
				continue
			}
			cur.body += line + "\n"
			continue
		}
		if strings.HasPrefix(line, "```") {
			cur = &block{lang: strings.TrimPrefix(line, "```"), heading: heading, index: count[heading], before: text.String()}
			count[heading]++
			text.Reset()
			continue
		}
		if strings.HasPrefix(line, "## ") {
			heading = strings.TrimPrefix(line, "## ")
		}
		text.WriteString(line + "\n")
	}
	return blocks
}

// shell runs the "$ " lines of the docs: crube (through cli.Run, in the
// current directory of the shell), and cd, export and unset.
type shell struct {
	t    *testing.T
	root string // where the shell starts
	cwd  string // absolute
	env  map[string]string
}

func newShell(t *testing.T, root string) *shell {
	return &shell{t: t, root: root, cwd: root, env: map[string]string{}}
}

// exec runs one command line and returns what it prints (standard output,
// then standard error). "; echo $?" at the end prints the exit code too.
func (s *shell) exec(line string) string {
	s.t.Helper()
	if i := strings.Index(line, "  #"); i >= 0 {
		line = strings.TrimRight(line[:i], " ")
	}
	line, echo := strings.CutSuffix(line, "; echo $?")
	words := splitWords(line)
	env := map[string]string{}
	for k, v := range s.env {
		env[k] = v
	}
	for len(words) > 0 && strings.Contains(words[0], "=") {
		k, v, _ := strings.Cut(words[0], "=")
		env[k] = v
		words = words[1:]
	}
	switch words[0] {
	case "cd":
		s.cwd = filepath.Join(s.cwd, filepath.FromSlash(words[1]))
		return ""
	case "export":
		k, v, _ := strings.Cut(words[1], "=")
		s.env[k] = v
		return ""
	case "unset":
		delete(s.env, words[1])
		return ""
	case "crube":
	default:
		s.t.Fatalf("the docs run %q, which the test shell does not know", line)
	}
	args := words[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") && args[0] != "help" {
		args = append([]string{args[0], "-C", s.cwd}, args[1:]...)
	}
	out, code := runCLI(args, env)
	if echo {
		out += fmt.Sprintf("%d\n", code)
	}
	return out
}

// runCLI runs crube with args, as the docs show it: not on a terminal, and
// built as release v0.1.0.
func runCLI(args []string, env map[string]string) (string, int) {
	var out bytes.Buffer
	code := cli.Run(args, cli.Env{
		Getenv:  func(k string) string { return env[k] },
		Stdout:  &out,
		Stderr:  &out,
		Version: "v0.1.0",
		Width:   func() int { return 0 },
	})
	return out.String(), code
}

// splitWords splits a command line at spaces; text in double quotes is one
// word.
func splitWords(line string) []string {
	var words []string
	var w strings.Builder
	quoted, inWord := false, false
	for _, r := range line {
		switch {
		case r == '"':
			quoted = !quoted
			inWord = true
		case r == ' ' && !quoted:
			if inWord {
				words = append(words, w.String())
				w.Reset()
				inWord = false
			}
		default:
			w.WriteRune(r)
			inWord = true
		}
	}
	if inWord {
		words = append(words, w.String())
	}
	return words
}

// runSession runs the "$ " lines of a block in s, and checks that each prints
// the lines below it in the docs. A line "..." stands for any lines.
func runSession(t *testing.T, s *shell, body string) {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		cmd, ok := strings.CutPrefix(lines[i], "$ ")
		if !ok {
			t.Fatalf("a line before the first command: %q", lines[i])
		}
		var want []string
		for i+1 < len(lines) && !strings.HasPrefix(lines[i+1], "$ ") {
			i++
			want = append(want, lines[i])
		}
		got := s.exec(cmd)
		if !matchLines(splitOutput(got), trimBlank(want)) {
			t.Errorf("$ %s\ngot:\n%s\nthe docs say:\n%s", cmd, got, strings.Join(want, "\n"))
		}
	}
}

func splitOutput(s string) []string {
	if s == "" {
		return nil
	}
	return trimBlank(strings.Split(strings.TrimSuffix(s, "\n"), "\n"))
}

// trimBlank drops the empty lines at the end.
func trimBlank(lines []string) []string {
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// matchLines reports whether got is want, where one line "..." in want
// stands for any lines.
func matchLines(got, want []string) bool {
	for i, w := range want {
		if w == "..." {
			rest := want[i+1:]
			return len(got) >= i+len(rest) && reflect.DeepEqual(got[:i], want[:i]) &&
				reflect.DeepEqual(got[len(got)-len(rest):], rest)
		}
	}
	return reflect.DeepEqual(got, want) || len(got) == 0 && len(want) == 0
}

// sameJSON reports whether two JSON texts hold the same value.
func sameJSON(t *testing.T, got, want string) bool {
	t.Helper()
	var g, w any
	if err := json.Unmarshal([]byte(got), &g); err != nil {
		t.Fatalf("%v: %s", err, got)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("the docs have JSON that does not parse: %v\n%s", err, want)
	}
	return reflect.DeepEqual(g, w)
}

// write makes the files under root: a name ending in "/" is a directory.
func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if strings.HasSuffix(name, "/") {
			if err := os.MkdirAll(p, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// project makes a project: the files, crube init, then the structure file.
func project(t *testing.T, files map[string]string, structure string) string {
	t.Helper()
	root := t.TempDir()
	write(t, root, files)
	if out, code := runCLI([]string{"init", "-C", root}, nil); code != 0 {
		t.Fatal(out)
	}
	write(t, root, map[string]string{".crube/structure.txt": structure})
	return root
}

// docLang is one language of the docs: the pages of English are X.md, those
// of Japanese X_ja.md. The examples of each are written with descriptions in
// that language, so each runs in a project described in it.
type docLang struct {
	suffix    string // added to the name of a page before ".md"
	structure string // the structure file of cliProject
	findWord  string // the word the find examples look for
	csvDesc   string // the desc of data/*.csv
	errors    string // the heading of the error examples in cli.md
	snapshot  string // the text before the structure file shown in getting-started.md
}

var (
	english = docLang{
		suffix: "",
		structure: "src/\tApplication source code\n" +
			"\tauth/\tAuthentication (login, sessions, token refresh)\n" +
			"\t\t| The entry point is index.ts.\n" +
			"\t\tsession.ts\tSession storage and expiry\n" +
			"\t\t\t| Sessions are kept in Redis; each access extends the expiry.\n" +
			"\t\t\t| The expiry is set by session.ttl in config/auth.yaml.\n" +
			"\t\t*_test.go\tTests of the file of the same name\n" +
			"\tmain.go\tEntry point\n" +
			"data/\tData files\n" +
			"\t*.csv\tCSV data\n",
		findWord: "session",
		csvDesc:  "CSV data",
		errors:   "Output conventions",
		snapshot: "At this point `.crube/structure.txt`",
	}
	japanese = docLang{
		suffix: "_ja",
		structure: "src/\tアプリケーションのソースコード\n" +
			"\tauth/\t認証(ログイン、セッション、トークン更新)\n" +
			"\t\t| 入口は index.ts。\n" +
			"\t\tsession.ts\tセッションの保存と有効期限の管理\n" +
			"\t\t\t| セッションは Redis に保存し、アクセスのたびに有効期限を延長する。\n" +
			"\t\t\t| 有効期限は config/auth.yaml の session.ttl で設定する。\n" +
			"\t\t*_test.go\t同名ファイルのテスト\n" +
			"\tmain.go\tエントリポイント\n" +
			"data/\tデータ格納場所\n" +
			"\t*.csv\tCSVデータ\n",
		findWord: "セッション",
		csvDesc:  "CSVデータ",
		errors:   "出力の共通の決まり",
		snapshot: "この時点の `.crube/structure.txt`",
	}
	docLangs = []docLang{english, japanese}
)

// page is the name of the page name ("docs/reference/cli.md") in l.
func (l docLang) page(name string) string {
	return strings.TrimSuffix(name, ".md") + l.suffix + ".md"
}

// cliProject is the project of the examples of docs/reference/cli.md and
// README.md, described in English. It tells whether the symbolic link
// "current" could be made (not always on Windows).
func cliProject(t *testing.T) (root string, link bool) {
	t.Helper()
	return cliProjectIn(t, english)
}

// cliProjectIn is cliProject described in the language of l.
func cliProjectIn(t *testing.T, l docLang) (root string, link bool) {
	t.Helper()
	root = project(t, map[string]string{
		"src/main.go":         "",
		"src/auth/session.ts": "",
		"src/db/conn.go":      "",
		"data/2026.csv":       "",
	}, l.structure)
	link = os.Symlink("releases/v2", filepath.Join(root, "current")) == nil
	return root, link
}

// statusProject is the project of the status examples: src/db/ is not
// described, src/legacy/ is gone, and with invalid, the desc of src/x.go is
// 300 columns wide.
func statusProject(t *testing.T, invalid bool) string {
	t.Helper()
	desc := "The X"
	if invalid {
		desc = strings.Repeat("x", 300)
	}
	return project(t, map[string]string{
		"src/x.go":       "",
		"src/db/conn.go": "",
	}, "src/\tSource\n\tx.go\t"+desc+"\n\tlegacy/\tOld code\n")
}
