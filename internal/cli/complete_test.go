package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// completeProject is a project with src/auth/ (described), src/db/ (not
// described), data/*.csv (a pattern) and build/ (ignored).
func completeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range []string{"src/main.go", "src/auth/session.ts", "src/db/conn.go", "data/2026.csv", "build/out.bin"} {
		full := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, ".crube"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"structure.txt":    "src/\tSource\n\tauth/\tAuthentication\n\tmain.go\tEntry point\ndata/\tData\n\t*.csv\tCSV data\nbuild/\tOutput\n",
		"structure_ja.txt": "src/\tソース\n",
		"ignore":           "build/\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, ".crube", name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestComplete(t *testing.T) {
	dir := completeProject(t)
	tests := []struct {
		lang string
		args []string
		want string
	}{
		// The first word.
		{"", []string{"--word="}, "init\tcreate .crube/\nls\tlist a directory with descriptions\n" +
			"tree\tshow the descriptions as a tree\nfind\tsearch paths and descriptions\n" +
			"show\tshow the description of one path\nstatus\treport untracked, orphan and invalid entries\n" +
			"set\tadd or change a description\nrm\tremove a description\nmv\tmove a description\n" +
			"mcp\trun as an MCP server over stdio\ncompletion\tprint a shell completion script\n"},
		{"", []string{"--word=s", "--"}, "show\tshow the description of one path\nstatus\treport untracked, orphan and invalid entries\nset\tadd or change a description\n"},
		{"", []string{"--word=-", "--"}, "--help\n--version\n"},
		{"", []string{"--word=", "--", "lss"}, ""},
		{"", []string{"--word=", "--", "--help"}, ""},

		// Options: those of the command not on the line yet, and --help.
		{"", []string{"--word=-", "--", "ls"}, "--full\n--json\n-C\n--no-ignore\n--help\n"},
		{"", []string{"--word=--", "--", "ls", "--json", "-h"}, "--full\n--no-ignore\n"},
		{"", []string{"--word=-", "--", "status", "--require=all", "-C=x"}, "--json\n--no-ignore\n--help\n"},
		{"", []string{"--word=-", "--", "init"}, "-C\n--help\n"},
		{"", []string{"--word=-", "--", "mcp"}, "--allow-write\n--help\n"},
		{"", []string{"--word=-", "--", "completion"}, "--help\n"},
		{"", []string{"--word=-", "--", "rm", "--"}, ""},

		// The values of options.
		{"", []string{"--word=", "--", "status", "--require"}, "all\ndirs\n"},
		{"", []string{"--word=d", "--", "status", "--require"}, "dirs\n"},
		{"", []string{"--word=a", "--", "status", "--require", "="}, "all\n"}, // as bash splits --require=a
		{"", []string{"--word=", "--", "tree", "-d"}, ""},
		{"", []string{"--word=", "--", "find", "--limit"}, ""},
		{"", []string{"--word=", "--", "set", "x", "--detail"}, ""},
		{"", []string{"--word=", "--", "ls", "-C"}, ""},
		{"", []string{"--word=", "--", "completion"}, "bash\nzsh\nfish\npowershell\n"},
		{"", []string{"--word=f", "--", "completion"}, "fish\n"},
		{"", []string{"--word=", "--", "completion", "bash"}, ""},

		// Entry paths, from where -C says.
		{"", []string{"--word=", "--", "tree", "-C", dir}, "src/\tSource\ndata/\tData\n"},
		{"", []string{"--word=src/", "--", "tree", "-C", dir}, "src/auth/\tAuthentication\nsrc/main.go\tEntry point\n"},
		{"", []string{"--word=src/", "--", "tree", "-C=" + dir}, "src/auth/\tAuthentication\nsrc/main.go\tEntry point\n"},
		{"", []string{"--word=src/", "--", "tree", "-C", "=", dir}, "src/auth/\tAuthentication\nsrc/main.go\tEntry point\n"},
		{"", []string{"--word=data/", "--", "rm", "-C", dir}, "data/*.csv\tCSV data\n"},
		{"", []string{"--word=", "--", "mv", "-C", dir, "src/"}, "src/\tSource\ndata/\tData\n"},
		{"", []string{"--word=", "--", "mv", "-C", dir, "src/", "data/"}, ""},
		{"", []string{"--word=", "--", "rm", "-C", dir, "--", "x"}, ""},
		{"", []string{"--word=-", "--", "rm", "-C", dir, "--"}, ""},
		{"", []string{"--word=src/", "--", "set", "-C", dir}, "src/auth/\tAuthentication\nsrc/main.go\tEntry point\nsrc/db/\n"},
		{"", []string{"--word=", "--", "set", "-C", dir, "src/"}, ""},
		{"", []string{"--word=", "--", "set", "-C", dir, "-"}, ""}, // a lone - is a path, as parse reads it
		{"", []string{"--word=", "--", "tree", "--no-ignore", "-C", dir}, "src/\tSource\ndata/\tData\nbuild/\tOutput\n"},
		{"ja", []string{"--word=", "--", "tree", "-C", dir}, "src/\tソース\n"},
		{"", []string{"--word=", "--", "ls", "-C", dir}, ""},
		{"", []string{"--word=", "--", "show", "-C", dir}, ""},
		{"", []string{"--word=", "--", "find", "-C", dir}, ""},

		// What goes wrong is no candidates.
		{"", []string{"--word=", "--", "tree", "-C", badProject(t)}, ""},
		{"", []string{"--word=", "--", "tree", "-C", t.TempDir()}, ""},
		{"ja_JP", []string{"--word=", "--", "tree", "-C", dir}, ""},
		{"", []string{"--word=a//", "--", "tree", "-C", dir}, ""},
	}
	for _, tt := range tests {
		code, out, errs := run(t, tt.lang, append([]string{"__complete"}, tt.args...)...)
		if code != 0 || out != tt.want || errs != "" {
			t.Errorf("%s %q: %d %q %q, want %q", tt.lang, tt.args, code, out, errs, tt.want)
		}
	}
}

// badProject is a project whose structure file does not parse.
func badProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".crube"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".crube", "structure.txt"), []byte("a/\n  b\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

// A desc is cut at 60 columns, for the menu of a shell: one of 60 is kept.
func TestCompleteCutsDesc(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".crube"), 0o755); err != nil {
		t.Fatal(err)
	}
	desc := strings.Repeat("あ", 31) // 62 columns
	fits := strings.Repeat("あ", 30) // 60 columns
	if err := os.WriteFile(filepath.Join(dir, ".crube", "structure.txt"), []byte("a\t"+desc+"\nb\t"+fits+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, out, errs := run(t, "", "__complete", "--word=", "--", "rm", "-C", dir)
	want := "a\t" + strings.Repeat("あ", 29) + "…\n" + // 58 columns and the 1 of …
		"b\t" + fits + "\n"
	if code != 0 || out != want || errs != "" {
		t.Errorf("%d %q %q, want %q", code, out, errs, want)
	}
}

func TestCompletionScripts(t *testing.T) {
	for _, shell := range shells {
		code, out, errs := run(t, "", "completion", shell)
		if code != 0 || errs != "" || !strings.Contains(out, "crube __complete \"--word=") {
			t.Errorf("%s: %d %q %q", shell, code, out, errs)
		}
	}
}
