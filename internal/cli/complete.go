package cli

import (
	"embed"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/amisonnet8/crube/internal/core"
	"github.com/amisonnet8/crube/internal/lang"
)

// Shell completion (docs/reference/cli.md "crube completion").

// The scripts are fixed text. They do not list the commands: at every TAB
// they ask crube __complete, which reads the tables of cli.go, so a script
// cannot disagree with the crube that is installed.
//
//go:embed completions
var completionScripts embed.FS

// shells are the shells that have a script, in the order they are named.
var shells = []string{"bash", "zsh", "fish", "powershell"}

var scriptFiles = map[string]string{
	"bash":       "completions/crube.bash",
	"zsh":        "completions/crube.zsh",
	"fish":       "completions/crube.fish",
	"powershell": "completions/crube.ps1",
}

// completeCommand is the command the scripts call. It is not in the usage:
// it is not for typing.
const completeCommand = "__complete"

// candidateWidth is how many columns of a desc a candidate has: it is for the
// menu of a shell.
const candidateWidth = 60

func (r *runner) completion(p *parsed) (int, error) {
	shell := p.args[0]
	if !slices.Contains(shells, shell) {
		return 0, r.argError(fmt.Errorf("unknown shell %q (use bash, zsh, fish or powershell)", shell))
	}
	script, err := completionScripts.ReadFile(scriptFiles[shell])
	if err != nil {
		return 0, err // the files are embedded, so this is a build without them
	}
	_, err = r.env.Stdout.Write(script)
	return exitOK, err
}

// candidate is one thing that can come next on a command line.
type candidate struct {
	value, desc string
}

// complete runs crube __complete [--word=<word>] -- <words>...: it lists what
// can come next after the words, of what starts with word. The scripts call
// it at every TAB, so it never complains: whatever the words are and whatever
// the project holds, it exits 0 and says nothing on Stderr.
func complete(args []string, env Env) int {
	var word string
	var words []string // the words after "--"
	for i, a := range args {
		if a == "--" {
			words = args[i+1:]
			break
		}
		if w, ok := strings.CutPrefix(a, "--word="); ok {
			word = w
		}
	}
	var b strings.Builder
	for _, c := range candidatesAt(walk(words), word, env) {
		b.WriteString(c.value)
		if c.desc != "" {
			b.WriteString("\t" + cut(c.desc, candidateWidth))
		}
		b.WriteString("\n")
	}
	_, _ = io.WriteString(env.Stdout, b.String()) // nowhere to report a failure
	return exitOK
}

// position is what a walk over the words of a line finds out about where the
// next word stands.
type position struct {
	name     string   // the command, "" before it
	cmd      *command // nil for a command that crube does not have
	seen     map[string]bool
	pending  string // an option that waits for its value
	args     int    // the positional arguments so far
	dashdash bool   // after "--": no option can come any more
	dir      string // the value of -C
	noIgnore bool
}

// walk reads the words the way parse does, but forgives everything: a line
// that is being typed is wrong more often than not.
func walk(words []string) position {
	p := position{seen: map[string]bool{}}
	for _, w := range words {
		switch {
		case p.pending != "":
			if w == "=" {
				continue // bash splits --require=all into three words
			}
			if p.pending == "-C" {
				p.dir = w
			}
			p.pending = ""
		case p.name == "":
			p.name = w
			p.cmd = commands[w]
		case w == "--" && !p.dashdash:
			p.dashdash = true
		case !p.dashdash && len(w) >= 2 && w[0] == '-':
			name, value, hasValue := strings.Cut(w, "=")
			if name == "-h" {
				name = "--help"
			}
			p.seen[name] = true
			switch {
			case name == "-C" && hasValue:
				p.dir = value
			case valueOptions[name] && !hasValue:
				p.pending = name
			case name == "--no-ignore":
				p.noIgnore = true
			}
		default:
			p.args++
		}
	}
	return p
}

// candidatesAt lists what can stand at a position, of what starts with what
// has been typed of the word.
func candidatesAt(p position, word string, env Env) []candidate {
	option := strings.HasPrefix(word, "-") && !p.dashdash
	var list []candidate
	switch {
	case p.name == "" && option:
		list = []candidate{{value: "--help"}, {value: "--version"}}
	case p.name == "":
		for _, name := range order {
			list = append(list, candidate{value: name, desc: commands[name].summary})
		}
	case p.cmd == nil:
	case p.pending == "--require":
		list = []candidate{{value: "all"}, {value: "dirs"}}
	case p.pending != "":
		// A number, a text, or a directory, which the shell knows about.
	case option:
		for _, name := range append(slices.Clone(p.cmd.options), "--help") {
			if !p.seen[name] {
				list = append(list, candidate{value: name})
			}
		}
	case p.name == "completion" && p.args == 0:
		for _, s := range shells {
			list = append(list, candidate{value: s})
		}
	case takesEntryPath(p.name, p.args):
		// core has chosen them by the word already, reading "\" as "/" on
		// Windows as an entry path is read.
		return entryCandidates(p, word, env)
	}
	return slices.DeleteFunc(list, func(c candidate) bool { return !strings.HasPrefix(c.value, word) })
}

// takesEntryPath says whether the positional argument at index i of a command
// is an entry path.
func takesEntryPath(name string, i int) bool {
	switch name {
	case "tree", "set", "rm":
		return i == 0
	case "mv":
		return i <= 1
	}
	return false
}

// entryCandidates are the entry paths that word can go on with, in the
// project the line is in. Anything that goes wrong (no .crube/, a structure
// file that does not parse) is no candidates: a message at every TAB would
// break the line being typed.
func entryCandidates(p position, word string, env Env) []candidate {
	proj, err := core.Open(core.Options{Dir: p.dir, Lang: env.Getenv(lang.EnvVar), NoIgnore: p.noIgnore})
	if err != nil {
		return nil
	}
	found, err := proj.Complete(word, p.name == "set")
	if err != nil {
		return nil
	}
	list := make([]candidate, len(found))
	for i, c := range found {
		list[i] = candidate{value: c.Path, desc: c.Desc}
	}
	return list
}
