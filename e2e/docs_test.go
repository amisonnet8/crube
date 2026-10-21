package e2e

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestGettingStarted runs docs/examples/getting-started.md from the top, in
// one shell, as a reader would.
func TestGettingStarted(t *testing.T) {
	for _, l := range docLangs {
		page := l.page("docs/examples/getting-started.md")
		t.Run(page, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, map[string]string{
				"myapp/go.mod":                  "",
				"myapp/cmd/myapp/main.go":       "",
				"myapp/internal/store/store.go": "",
			})
			s := newShell(t, dir)
			snapshots := 0
			for _, b := range readBlocks(t, page) {
				switch {
				case strings.HasPrefix(b.body, "$ "):
					runSession(t, s, b.body)
				case strings.Contains(b.before, l.snapshot):
					snapshots++
					got, err := os.ReadFile(filepath.Join(dir, "myapp", ".crube", "structure.txt"))
					if err != nil {
						t.Fatal(err)
					}
					if string(got) != b.body {
						t.Errorf(".crube/structure.txt:\n%s\nthe docs say:\n%s", got, b.body)
					}
				}
			}
			if snapshots != 1 {
				t.Errorf("%d blocks after %q, want 1", snapshots, l.snapshot)
			}
		})
	}
}

// TestCLIExamples runs the command examples of docs/reference/cli.md, each
// block from the project root of the project it is about.
func TestCLIExamples(t *testing.T) {
	for _, l := range docLangs {
		page := l.page("docs/reference/cli.md")
		for _, b := range readBlocks(t, page) {
			if !strings.HasPrefix(b.body, "$ ") {
				continue
			}
			t.Run(page+"/"+b.heading+"/"+strings.SplitN(b.body, "\n", 2)[0], func(t *testing.T) {
				var root string
				switch {
				case b.heading == "crube status" && strings.Contains(b.body, "invalid "):
					root = statusProject(t, true)
				case b.heading == "crube status":
					root = project(t, map[string]string{"a.txt": ""}, "a.txt\tA\n")
				default:
					var link bool
					root, link = cliProjectIn(t, l)
					if !link && strings.Contains(b.body, "crube show current") {
						t.Skip("cannot make a symbolic link here")
					}
				}
				runSession(t, newShell(t, root), b.body)
			})
		}
	}
}

// TestREADMEExample runs the example at the top of README.md and of
// docs/README.md, in the project of cli.md.
func TestREADMEExample(t *testing.T) {
	for _, l := range docLangs {
		for _, file := range []string{l.page("README.md"), l.page("docs/README.md")} {
			t.Run(file, func(t *testing.T) {
				root, _ := cliProjectIn(t, l)
				n := 0
				for _, b := range readBlocks(t, file) {
					if strings.HasPrefix(b.body, "$ ") {
						runSession(t, newShell(t, root), b.body)
						n++
					}
				}
				if n == 0 {
					t.Errorf("%s has no example to run", file)
				}
			})
		}
	}
}

// jsonExample is how to get the output a JSON example of the docs shows.
type jsonExample struct {
	project func(t *testing.T, l docLang) string
	dir     string // where to run, from the project root
	args    func(l docLang) []string
}

func cliRoot(t *testing.T, l docLang) string {
	root, _ := cliProjectIn(t, l)
	return root
}

func statusRoot(invalid bool) func(t *testing.T, l docLang) string {
	return func(t *testing.T, _ docLang) string { return statusProject(t, invalid) }
}

func fixed(args ...string) func(docLang) []string {
	return func(docLang) []string { return args }
}

// TestJSONExamples checks every JSON example of the docs, in each language,
// against the output of the command it shows. An example is known by its page
// and its place among the JSON blocks of the page ("cli.md#2" is the third),
// the same in both languages. An example the test does not know fails, so
// that a new example gets checked too.
func TestJSONExamples(t *testing.T) {
	examples := map[string]jsonExample{
		"docs/reference/cli.md#0": {cliRoot, "", fixed("ls", "src", "--json")},
		"docs/reference/cli.md#1": {cliRoot, "src/auth", fixed("tree", "src/", "-d", "2", "--json")},
		"docs/reference/cli.md#2": {cliRoot, "", func(l docLang) []string { return []string{"find", l.findWord, "--json"} }},
		"docs/reference/cli.md#3": {cliRoot, "", fixed("show", "data/2026.csv", "--json")},
		"docs/reference/cli.md#4": {statusRoot(true), "", fixed("status", "--json")},
		"docs/reference/cli.md#5": {cliRoot, "", func(l docLang) []string { return []string{"set", "data/*.csv", l.csvDesc, "--json"} }},
		// The output of status --json in ai-workflow.md.
		"docs/examples/ai-workflow.md#0": {statusRoot(false), "", fixed("status", "--json")},
	}
	// JSON in the docs that is not the output of a command: the settings of
	// an MCP client.
	notOutput := map[string]bool{
		"docs/reference/mcp.md#0": true,
	}

	for _, l := range docLangs {
		seen := map[string]bool{}
		for _, doc := range []string{"docs/reference/cli.md", "docs/reference/mcp.md", "docs/examples/ai-workflow.md"} {
			n := 0
			for _, b := range readBlocks(t, l.page(doc)) {
				if b.lang != "json" {
					continue
				}
				key := doc + "#" + strconv.Itoa(n)
				n++
				seen[key] = true
				if notOutput[key] {
					continue
				}
				ex, ok := examples[key]
				if !ok {
					t.Errorf("%s (%s): a JSON example the test does not know; add how to get it to TestJSONExamples", key, l.page(doc))
					continue
				}
				t.Run(l.page(doc)+"#"+b.heading, func(t *testing.T) {
					root := ex.project(t, l)
					args := ex.args(l)
					got, _ := runCLI(append([]string{args[0], "-C", filepath.Join(root, filepath.FromSlash(ex.dir))}, args[1:]...), nil)
					if !sameJSON(t, got, b.body) {
						t.Errorf("crube %s\ngot:\n%s\nthe docs say:\n%s", strings.Join(args, " "), got, b.body)
					}
				})
			}
		}
		for key := range examples {
			if !seen[key] {
				t.Errorf("%s: no such JSON example in the %q pages any more", key, l.suffix)
			}
		}
	}
}

// TestErrorExamples checks that each error line in "Output conventions" of
// cli.md (and its Japanese page) is what crube prints in some case.
func TestErrorExamples(t *testing.T) {
	type errorCase struct {
		project func(t *testing.T) string // nil: no project
		args    []string                  // "-C <root>" is added after the command when project is set
	}
	root := func(t *testing.T) string { return cliRoot(t, english) }
	cases := map[string]errorCase{
		// The bad line is the 12th: 11 good lines, then one indented with spaces.
		"crube: .crube/structure.txt:12: indentation must be tabs": {func(t *testing.T) string {
			return project(t, nil, linesThenBad(11))
		}, []string{"tree"}},
		"crube: no .crube/ found (run crube init)":    {func(t *testing.T) string { return t.TempDir() }, []string{"ls"}},
		"crube: desc is 300 columns wide (limit 256)": {root, []string{"set", "a.txt", strings.Repeat("x", 300)}},
		"crube: src/foo.go: no such file":             {root, []string{"show", "src/foo.go"}},
		"crube: build/x: ignored by .crube/ignore (use --no-ignore)": {func(t *testing.T) string {
			r := root(t)
			write(t, r, map[string]string{"build/x": "", ".crube/ignore": "build/\n"})
			return r
		}, []string{"show", "build/x"}},
		"crube: .crube: .crube/ is always left out, even with --no-ignore":                                  {root, []string{"ls", ".crube", "--no-ignore"}},
		`crube: unknown command "lss" (see crube --help)`:                                                   {nil, []string{"lss"}},
		"crube: ls: unknown option --foo":                                                                   {nil, []string{"ls", "--foo"}},
		"crube: ls: -C needs a value":                                                                       {nil, []string{"ls", "-C"}},
		`crube: find: invalid --limit "x" (use a number of 1 or more)`:                                      {root, []string{"find", "--limit", "x", "q"}},
		"crube: ls: too many arguments (usage: crube ls [path] [--full] [--json] [-C <dir>] [--no-ignore])": {nil, []string{"ls", "a", "b"}},
		"crube: set: missing <desc> (usage: crube set <path> <desc> [--detail <text>] [--json] [-C <dir>])": {nil, []string{"set", "x"}},
	}

	for _, l := range docLangs {
		page := l.page("docs/reference/cli.md")
		n := 0
		for _, b := range readBlocks(t, page) {
			if b.heading != l.errors || strings.HasPrefix(b.body, "$ ") {
				continue
			}
			for _, line := range strings.Split(strings.TrimSuffix(b.body, "\n"), "\n") {
				n++
				c, ok := cases[line]
				if !ok {
					t.Errorf("%s: an error example the test does not know: %q", page, line)
					continue
				}
				args := c.args
				if c.project != nil {
					args = append([]string{args[0], "-C", c.project(t)}, args[1:]...)
				}
				got, code := runCLI(args, nil)
				if code != 2 || got != line+"\n" {
					t.Errorf("crube %s: %d %q, the docs say %q", strings.Join(c.args, " "), code, got, line)
				}
			}
		}
		if n != len(cases) {
			t.Errorf("%s has %d error examples, the test knows %d", page, n, len(cases))
		}
	}
}

// TestDocPairs checks that README.md and every page of docs/ come in pairs,
// X.md in English and X_ja.md in Japanese, that each starts with the switch to
// the other language, and that each links only to pages of its own language.
func TestDocPairs(t *testing.T) {
	pages := map[string]bool{"README.md": true, "README_ja.md": true}
	err := filepath.WalkDir(filepath.Join("..", "docs"), func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".md") {
			pages[filepath.ToSlash(strings.TrimPrefix(p, ".."+string(filepath.Separator)))] = true
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	link := regexp.MustCompile(`\]\(([^)#:]+\.md)(#[^)]*)?\)`)
	for p := range pages {
		ja := strings.HasSuffix(p, "_ja.md")
		var other, first string
		if ja {
			other = strings.TrimSuffix(p, "_ja.md") + ".md"
			first = "*[English](" + path.Base(other) + ") | **日本語***"
		} else {
			other = japanese.page(p)
			first = "*[日本語](" + path.Base(other) + ") | **English***"
		}
		if !pages[other] {
			t.Errorf("%s has no pair %s", p, other)
		}
		data, err := os.ReadFile(filepath.Join("..", filepath.FromSlash(p)))
		if err != nil {
			t.Fatal(err)
		}
		head, rest, _ := strings.Cut(string(data), "\n")
		if head != first {
			t.Errorf("%s starts with %q, want %q", p, head, first)
		}
		for _, m := range link.FindAllStringSubmatch(rest, -1) {
			if strings.HasSuffix(m[1], "_ja.md") != ja {
				t.Errorf("%s links to %s, a page of the other language", p, m[1])
			}
		}
	}
}

// linesThenBad is a structure file of n good lines (each its own entry) and
// a line indented with spaces after them.
func linesThenBad(n int) string {
	var b strings.Builder
	for i := range n {
		b.WriteString("f" + strconv.Itoa(i) + "\tF\n")
	}
	b.WriteString("  bad\n")
	return b.String()
}
