package search

import (
	"errors"
	"reflect"
	"testing"

	"github.com/amisonnet8/crube/internal/ignore"
	"github.com/amisonnet8/crube/internal/structure"
)

const sample = "src/\tアプリケーションのソースコード\n" +
	"\tauth/\t認証(ログイン、セッション、トークン更新)\n" +
	"\t\t| 入口は index.ts。\n" +
	"\t\tsession.ts\tセッションの保存と有効期限の管理\n" +
	"\t\t*_test.go\t同名ファイルのテスト\n" +
	"data/\tデータ格納場所\n" +
	"\t*.csv\tCSVデータ\n" +
	"\t\t| Exported from the Billing System every night.\n" +
	"node_modules/\tDependencies\n" +
	"\tleft-pad/\tSession helper, but ignored\n" +
	"docs/\n" +
	"\tREADME.md\tHow to read the docs\n"

func find(t *testing.T, query string, limit int, ign *ignore.Matcher) *Found {
	t.Helper()
	tree, err := structure.Parse([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	f, err := Find(tree, query, limit, ign)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func paths(f *Found) []string {
	var ps []string
	for _, r := range f.Items {
		ps = append(ps, r.Path)
	}
	return ps
}

func TestFind(t *testing.T) {
	ign := ignore.Parse([]byte("node_modules/\n"))
	tests := []struct {
		query string
		want  []string
	}{
		// The example of docs/reference/cli.md.
		{"セッション", []string{"src/auth/", "src/auth/session.ts"}},
		// By path, including the pattern of an entry.
		{"session.ts", []string{"src/auth/session.ts"}},
		{"*.csv", []string{"data/*.csv"}},
		{"src/auth", []string{"src/auth/", "src/auth/session.ts", "src/auth/*_test.go"}},
		// By detail.
		{"index.ts", []string{"src/auth/"}},
		// Case is ignored, in the query and in the text.
		{"billing system", []string{"data/*.csv"}},
		{"CSV", []string{"data/*.csv"}},
		{"readme", []string{"docs/README.md"}},
		// Words are ANDed; each may be in a different field.
		{"auth 有効期限", []string{"src/auth/session.ts"}},
		{"  data   csv  ", []string{"data/*.csv"}},
		{"セッション billing", nil},
		// An entry without a desc is found by its path.
		{"docs/", []string{"docs/", "docs/README.md"}},
		// Ignored entries, and what is below them, are not found.
		{"dependencies", nil},
		{"helper", nil},
		{"nothing-like-this", nil},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			f := find(t, tt.query, 0, ign)
			if got := paths(f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find(%q) = %q, want %q", tt.query, got, tt.want)
			}
			if f.More != 0 || f.Query != tt.query || f.Items == nil {
				t.Errorf("Found = %+v", f)
			}
		})
	}
}

func TestFindGivesDescAndDetail(t *testing.T) {
	f := find(t, "index.ts", 0, ignore.None())
	want := []Result{{Path: "src/auth/", Desc: "認証(ログイン、セッション、トークン更新)", Detail: "入口は index.ts。"}}
	if !reflect.DeepEqual(f.Items, want) {
		t.Errorf("items = %+v, want %+v", f.Items, want)
	}
}

func TestFindLimit(t *testing.T) {
	// "src" is in the paths of the four entries under src/ (and in nothing else).
	tests := []struct {
		limit int
		want  []string
		more  int
	}{
		{0, []string{"src/", "src/auth/", "src/auth/session.ts", "src/auth/*_test.go"}, 0},
		{-1, []string{"src/", "src/auth/", "src/auth/session.ts", "src/auth/*_test.go"}, 0},
		{2, []string{"src/", "src/auth/"}, 2},
		{4, []string{"src/", "src/auth/", "src/auth/session.ts", "src/auth/*_test.go"}, 0},
		{10, []string{"src/", "src/auth/", "src/auth/session.ts", "src/auth/*_test.go"}, 0},
	}
	for _, tt := range tests {
		f := find(t, "src/", tt.limit, ignore.None())
		if got := paths(f); !reflect.DeepEqual(got, tt.want) || f.More != tt.more {
			t.Errorf("limit %d: got %q more %d, want %q more %d", tt.limit, got, f.More, tt.want, tt.more)
		}
	}
}

func TestFindWithoutIgnore(t *testing.T) {
	f := find(t, "helper", 0, ignore.None())
	if got := paths(f); !reflect.DeepEqual(got, []string{"node_modules/left-pad/"}) {
		t.Errorf("got %q", got)
	}
}

func TestEmptyQuery(t *testing.T) {
	tree := &structure.Tree{}
	for _, q := range []string{"", "   ", "\t"} {
		if _, err := Find(tree, q, 0, ignore.None()); !errors.Is(err, ErrEmptyQuery) {
			t.Errorf("Find(%q) = %v, want ErrEmptyQuery", q, err)
		}
	}
}
