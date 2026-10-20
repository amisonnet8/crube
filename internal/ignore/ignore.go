// Package ignore evaluates .crube/ignore, the paths that ls, tree, find and
// status leave out (docs/reference/file-format.md "ignore").
//
// The format is that of .gitignore, written with the standard library only;
// .gitignore itself is never read. .crube/ at the project root is always left
// out, even with --no-ignore.
package ignore

import (
	"path"
	"strings"

	"github.com/amisonnet8/crube/internal/project"
)

// FileName is the name of the ignore file in .crube/.
const FileName = "ignore"

// Default is the ignore file that crube init writes: directories that hold
// the data of version control and package managers in almost any project.
const Default = `# Paths that crube leaves out of ls, tree, find and status.
# The format is that of .gitignore, but .gitignore itself is not read.
# .crube/ is always left out.
.git/
.hg/
.svn/
node_modules/
`

// Matcher tells whether a path is left out.
type Matcher struct {
	rules []rule
}

type rule struct {
	segs    []string // pattern split at "/"; "**" is any number of directories
	negate  bool     // "!": a match takes the path back in
	dirOnly bool     // trailing "/": matches directories only
}

// Parse reads the content of an ignore file. Lines that are not valid
// patterns are skipped, like git does.
func Parse(data []byte) *Matcher {
	m := &Matcher{}
	for _, line := range strings.Split(string(data), "\n") {
		if r, ok := parseLine(strings.TrimSuffix(line, "\r")); ok {
			m.rules = append(m.rules, r)
		}
	}
	return m
}

// None is the matcher for --no-ignore: it leaves out only .crube/.
func None() *Matcher {
	return &Matcher{}
}

func parseLine(s string) (rule, bool) {
	// Trailing spaces are dropped unless escaped with a backslash.
	for strings.HasSuffix(s, " ") && !strings.HasSuffix(s, `\ `) {
		s = s[:len(s)-1]
	}
	if s == "" || strings.HasPrefix(s, "#") {
		return rule{}, false
	}
	var r rule
	if strings.HasPrefix(s, "!") {
		r.negate = true
		s = s[1:]
	}
	if strings.HasSuffix(s, "/") {
		r.dirOnly = true
		s = strings.TrimSuffix(s, "/")
	}
	// A slash at the start or in the middle ties the pattern to the project
	// root; without one, the pattern matches a name at any depth.
	anchored := strings.Contains(s, "/")
	s = strings.TrimPrefix(s, "/")
	if s == "" {
		return rule{}, false
	}
	r.segs = strings.Split(s, "/")
	if !anchored {
		r.segs = append([]string{"**"}, r.segs...)
	}
	return r, true
}

// Ignored reports whether the path, given as its names from the project root,
// is left out. All names but the last are directories; isDir tells whether
// the last one is. A path is left out also when a directory above it is: a
// "!" line cannot take back a path inside a directory that is left out.
func (m *Matcher) Ignored(names []string, isDir bool) bool {
	if len(names) > 0 && names[0] == project.DirName {
		return true
	}
	for i := 1; i <= len(names); i++ {
		if m.matches(names[:i], isDir || i < len(names)) {
			return true
		}
	}
	return false
}

// matches applies the rules to this one path: the last rule that matches it
// decides.
func (m *Matcher) matches(names []string, isDir bool) bool {
	ignored := false
	for _, r := range m.rules {
		if r.dirOnly && !isDir {
			continue
		}
		if matchSegs(r.segs, names) {
			ignored = !r.negate
		}
	}
	return ignored
}

// matchSegs matches pattern segments against path names. "**" matches any
// number of names; at the end of a pattern it needs at least one ("a/**" is
// what is inside a, not a itself). Other segments are matched with
// path.Match, which also handles the backslash escapes (\#, \!, "\ ").
func matchSegs(segs, names []string) bool {
	for len(segs) > 0 {
		if segs[0] == "**" {
			rest := segs[1:]
			if len(rest) == 0 {
				return len(names) > 0
			}
			for i := 0; i <= len(names); i++ {
				if matchSegs(rest, names[i:]) {
					return true
				}
			}
			return false
		}
		if len(names) == 0 {
			return false
		}
		if ok, err := path.Match(segs[0], names[0]); err != nil || !ok {
			return false
		}
		segs, names = segs[1:], names[1:]
	}
	return len(names) == 0
}
