package match

import (
	"reflect"
	"strings"
	"testing"

	"github.com/amisonnet8/crube/internal/structure"
)

// The structure used by the tests. Comments say which rule each entry is for.
const sample = "src/\tSource\n" +
	"\tauth/\tAuthentication\n" + // exact, before the pattern */
	"\t\tsession.ts\tSession\n" +
	"\t\t*_test.go\tTests\n" +
	"\t\t[id].tsx\tDynamic route\n" + // a name with pattern characters
	"\t*.go\tGo files\n" + // written before *_test.go: wins for x_test.go
	"\t*_test.go\tNever reached for .go files\n" +
	"\t*/\tAny other directory\n" +
	"\t\tmain.go\tEntry point of a command\n" +
	"\t[\tA malformed pattern\n" +
	"data/\n" +
	"\t*.csv\tCSV data\n" +
	"*.md\tMarkdown\n" +
	"README.md\tRead me first\n" + // exact, though *.md is written above
	"x\tA file named x\n"

func parse(t *testing.T) []*structure.Entry {
	t.Helper()
	tree, err := structure.Parse([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	return tree.Entries
}

func TestFind(t *testing.T) {
	entries := parse(t)
	tests := []struct {
		path  string // a trailing "/" means a directory
		entry string // the path of the matched entry; "" for none
	}{
		{"src/", "src/"},
		{"src/auth/", "src/auth/"},
		{"src/auth/session.ts", "src/auth/session.ts"},
		{"src/auth/login_test.go", "src/auth/*_test.go"},
		{"src/auth/[id].tsx", "src/auth/[id].tsx"},
		{"src/auth/i.tsx", "src/auth/[id].tsx"}, // as a pattern, [id] is one of i and d
		{"src/auth/x.tsx", ""},
		{"src/main.go", "src/*.go"},
		{"src/main_test.go", "src/*.go"}, // the first pattern written wins
		{"src/cmd/", "src/*/"},
		{"src/x.go/", "src/*/"}, // a directory: *.go is only for files
		{"src/cmd/main.go", "src/*/main.go"},
		{"src/cmd/other.go", ""}, // the children of */ have no pattern for it
		{"src/[", "src/["},       // a malformed pattern still matches its own name exactly
		{"src/a", ""},
		{"data/", "data/"},
		{"data/2026.csv", "data/*.csv"},
		{"data/sub/", ""},
		{"data/sub/x.csv", ""}, // a pattern matches only names directly in its directory
		{"README.md", "README.md"},
		{"CHANGES.md", "*.md"},
		{"x", "x"},
		{"x/", ""},     // the entry x is for files only
		{"src", ""},    // the entry src/ is for directories only
		{"other/", ""}, // no entry at all
		{"other/a.md", ""},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			names, isDir := split(tt.path)
			m := Find(entries, names, isDir)
			switch {
			case tt.entry == "" && m != nil:
				t.Errorf("matched %q, want none", m.Path)
			case tt.entry != "" && m == nil:
				t.Errorf("matched none, want %q", tt.entry)
			case m != nil && m.Path != tt.entry:
				t.Errorf("matched %q, want %q", m.Path, tt.entry)
			}
		})
	}
}

func TestFindGivesTheEntry(t *testing.T) {
	m := Find(parse(t), []string{"data", "2026.csv"}, false)
	if m == nil || m.Entry.Name != "*.csv" || m.Entry.Desc != "CSV data" {
		t.Fatalf("got %+v", m)
	}
	if Find(parse(t), nil, true) != nil {
		t.Error("the empty path (the project root) matched an entry")
	}
}

func TestWalk(t *testing.T) {
	entries := parse(t)
	tests := []struct {
		path string
		want []string // names of the entries matched
	}{
		{"src/auth/session.ts", []string{"src/", "auth/", "session.ts"}},
		{"src/auth/deep/er/", []string{"src/", "auth/"}},
		{"src/cmd/x/", []string{"src/", "*/"}},
		{"data/sub/x.csv", []string{"data/"}},
		{"other/", nil},
	}
	for _, tt := range tests {
		names, isDir := split(tt.path)
		var got []string
		for _, e := range Walk(entries, names, isDir) {
			got = append(got, e.Name)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Walk(%s) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestEntryPath(t *testing.T) {
	chain := Walk(parse(t), []string{"src", "auth", "session.ts"}, false)
	if got, want := EntryPath(chain), "src/auth/session.ts"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := EntryPath(nil); got != "" {
		t.Errorf("empty chain: got %q", got)
	}
}

// split turns "a/b/" into the names ["a" "b"] and isDir true.
func split(p string) ([]string, bool) {
	isDir := strings.HasSuffix(p, "/")
	return strings.Split(strings.TrimSuffix(p, "/"), "/"), isDir
}
