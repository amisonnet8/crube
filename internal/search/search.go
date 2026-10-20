// Package search finds entries of a structure file by their path, desc and
// detail (docs/reference/cli.md "crube find"). There is no index: the caller
// reads the structure file each time and passes it in.
package search

import (
	"errors"
	"strings"

	"github.com/amisonnet8/crube/internal/ignore"
	"github.com/amisonnet8/crube/internal/structure"
)

// ErrEmptyQuery is returned for a query with no word in it.
var ErrEmptyQuery = errors.New("empty query")

// Result is one entry found.
type Result struct {
	Path   string `json:"path"` // the entry path, such as "data/*.csv"
	Desc   string `json:"desc"`
	Detail string `json:"detail"`
}

// Found is the result of Find.
type Found struct {
	Query string   `json:"query"`
	More  int      `json:"more"` // how many were left out by the limit
	Items []Result `json:"items"`
}

// Find returns the entries of tree that hold every word of query (split at
// white space) somewhere in their path, desc or detail, ignoring case, in the
// order of the structure file. Entries whose path ign leaves out are not
// found. With limit > 0, at most limit entries are returned and More counts
// the rest.
func Find(tree *structure.Tree, query string, limit int, ign *ignore.Matcher) (*Found, error) {
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return nil, ErrEmptyQuery
	}
	f := &Found{Query: query, Items: []Result{}}
	var walk func(entries []*structure.Entry, names []string)
	walk = func(entries []*structure.Entry, names []string) {
		for _, e := range entries {
			p := append(names[:len(names):len(names)], strings.TrimSuffix(e.Name, "/"))
			if ign.Ignored(p, e.IsDir()) {
				continue
			}
			path := strings.Join(p[:len(p)-1], "/")
			if path != "" {
				path += "/"
			}
			path += e.Name
			if hasAll(strings.ToLower(path+"\n"+e.Desc+"\n"+e.Detail), words) {
				if limit > 0 && len(f.Items) >= limit {
					f.More++
				} else {
					f.Items = append(f.Items, Result{Path: path, Desc: e.Desc, Detail: e.Detail})
				}
			}
			walk(e.Children, p)
		}
	}
	walk(tree.Entries, nil)
	return f, nil
}

func hasAll(text string, words []string) bool {
	for _, w := range words {
		if !strings.Contains(text, w) {
			return false
		}
	}
	return true
}
