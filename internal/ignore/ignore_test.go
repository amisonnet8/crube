package ignore

import (
	"strings"
	"testing"
)

// The expected results were taken from git itself: a repository with the
// patterns as its .gitignore, the path made as a file or a directory, and
// `git check-ignore -q <path>` (without a trailing slash).
func TestIgnoredLikeGit(t *testing.T) {
	tests := []struct {
		patterns []string
		path     string // a trailing "/" means a directory
		want     bool
	}{
		{[]string{"*.log"}, "a.log", true},
		{[]string{"*.log"}, "sub/a.log", true},
		{[]string{"*.log", "!keep.log"}, "keep.log", false},
		{[]string{"*.log", "!keep.log"}, "sub/keep.log", false},
		{[]string{"!keep.log", "*.log"}, "keep.log", true},
		{[]string{"build/"}, "build/", true},
		{[]string{"build/"}, "build", false},
		{[]string{"build/"}, "src/build/", true},
		{[]string{"build/"}, "build/x.txt", true},
		{[]string{"build/", "!build/keep.txt"}, "build/keep.txt", true},
		{[]string{"/root.txt"}, "root.txt", true},
		{[]string{"/root.txt"}, "sub/root.txt", false},
		{[]string{"docs/*.tmp"}, "docs/a.tmp", true},
		{[]string{"docs/*.tmp"}, "x/docs/a.tmp", false},
		{[]string{"docs/*.tmp"}, "docs/sub/a.tmp", false},
		{[]string{"**/cache"}, "cache/", true},
		{[]string{"**/cache"}, "a/b/cache", true},
		{[]string{"a/**"}, "a/", false},
		{[]string{"a/**"}, "a/b/c.txt", true},
		{[]string{"x/**/y"}, "x/y", true},
		{[]string{"x/**/y"}, "x/p/q/y", true},
		{[]string{"x/**/y"}, "z/x/y", false},
		{[]string{`\#hash`}, "#hash", true},
		{[]string{"#hash"}, "#hash", false},
		{[]string{`\!bang`}, "!bang", true},
		{[]string{`trail\ `}, "trail ", true},
		{[]string{"foo  "}, "foo", true},
		{[]string{"f?o"}, "fxo", true},
		{[]string{"[ab].txt"}, "b.txt", true},
		{[]string{"sub/inner/"}, "sub/inner/", true},
		{[]string{"sub/inner/"}, "x/sub/inner/", false},
		{[]string{"*"}, "anything", true},
		{[]string{"", "   ", "# c"}, "a", false},
		{[]string{"!*.log"}, "a.log", false},
		{[]string{"a/**/"}, "a/b/", true},
		{[]string{"a/**/"}, "a/b", false},
		{[]string{"**"}, "x/y", true},
		{[]string{"src/"}, "src/a/b.txt", true},
		{[]string{"b"}, "a/b/c.txt", true},
		{[]string{"/a/b"}, "a/b/", true},
	}
	for _, tt := range tests {
		name := strings.Join(tt.patterns, ",") + " " + tt.path
		t.Run(name, func(t *testing.T) {
			m := Parse([]byte(strings.Join(tt.patterns, "\n") + "\n"))
			names, isDir := split(tt.path)
			if got := m.Ignored(names, isDir); got != tt.want {
				t.Errorf("Ignored(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestCrubeAlwaysIgnored(t *testing.T) {
	for _, m := range []*Matcher{None(), Parse(nil), Parse([]byte("!.crube/\n!.crube/**\n"))} {
		for _, p := range []string{".crube/", ".crube/structure.txt", ".crube/ignore"} {
			names, isDir := split(p)
			if !m.Ignored(names, isDir) {
				t.Errorf("%s is not ignored", p)
			}
		}
		// Only the .crube/ of the project root is crube's own.
		if names, isDir := split("sub/.crube/"); m.Ignored(names, isDir) {
			t.Error("sub/.crube/ is ignored")
		}
	}
}

func TestNoneIgnoresNothingElse(t *testing.T) {
	for _, p := range []string{".git/", "node_modules/x.js", "a.log"} {
		names, isDir := split(p)
		if None().Ignored(names, isDir) {
			t.Errorf("None ignores %s", p)
		}
	}
}

func TestCRLF(t *testing.T) {
	m := Parse([]byte("*.log\r\nbuild/\r\n"))
	if !m.Ignored([]string{"a.log"}, false) || !m.Ignored([]string{"build"}, true) {
		t.Error("a CRLF ignore file is not read")
	}
}

func TestMalformedPattern(t *testing.T) {
	m := Parse([]byte("[\n*.log\n"))
	if m.Ignored([]string{"["}, false) {
		t.Error("a malformed pattern matched")
	}
	if !m.Ignored([]string{"a.log"}, false) {
		t.Error("a malformed line stopped the lines after it")
	}
}

func TestDefault(t *testing.T) {
	m := Parse([]byte(Default))
	for p, want := range map[string]bool{
		".git/":             true,
		".git/config":       true,
		".hg/":              true,
		".svn/":             true,
		"node_modules/":     true,
		"web/node_modules/": true,
		".gitignore":        false,
		".github/":          false,
		"src/main.go":       false,
		"docs/node_modules": false, // a file of that name
		".crube/":           true,
	} {
		names, isDir := split(p)
		if got := m.Ignored(names, isDir); got != want {
			t.Errorf("Default: Ignored(%s) = %v, want %v", p, got, want)
		}
	}
}

func split(p string) ([]string, bool) {
	isDir := strings.HasSuffix(p, "/")
	return strings.Split(strings.TrimSuffix(p, "/"), "/"), isDir
}
