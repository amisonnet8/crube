package cli

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// valueOptions are the options that take a value: "-C dir" or "-C=dir". Every
// other option is a flag.
var valueOptions = map[string]bool{
	"-C":        true,
	"-d":        true,
	"--limit":   true,
	"--detail":  true,
	"--require": true,
}

// parsed is a command line after the command name.
type parsed struct {
	opts map[string]string // the options given, with their values ("" for a flag)
	args []string          // the positional arguments, in order
	help bool              // -h or --help was given
}

// has reports whether the option was given.
func (p *parsed) has(name string) bool {
	_, ok := p.opts[name]
	return ok
}

// parse reads args for a command that accepts the options allowed. Options
// may come before or after the positional arguments; after "--" everything is
// positional. A value option takes the next argument as it is, even one that
// starts with "-" (so --detail "" and --detail -x work).
func parse(args []string, allowed []string) (*parsed, error) {
	p := &parsed{opts: map[string]string{}}
	rest := args
	for len(rest) > 0 {
		a := rest[0]
		rest = rest[1:]
		if a == "--" {
			p.args = append(p.args, rest...)
			break
		}
		if len(a) < 2 || a[0] != '-' {
			p.args = append(p.args, a)
			continue
		}
		if a == "-h" || a == "--help" {
			p.help = true
			continue
		}
		name, value, hasValue := strings.Cut(a, "=")
		if !slices.Contains(allowed, name) {
			return nil, fmt.Errorf("unknown option %s", name)
		}
		if !valueOptions[name] {
			if hasValue {
				return nil, fmt.Errorf("%s takes no value", name)
			}
			p.opts[name] = ""
			continue
		}
		if !hasValue {
			if len(rest) == 0 {
				return nil, fmt.Errorf("%s needs a value", name)
			}
			value = rest[0]
			rest = rest[1:]
		}
		p.opts[name] = value
	}
	return p, nil
}

// count reads the value of an option that is a number of 1 or more; 0 when
// the option is not given.
func (p *parsed) count(name string) (int, error) {
	v, ok := p.opts[name]
	if !ok {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("invalid %s %q (use a number of 1 or more)", name, v)
	}
	return n, nil
}
