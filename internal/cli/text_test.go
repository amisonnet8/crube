package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runText runs the command line with the terminal width given (0: not a
// terminal).
func runText(t *testing.T, width int, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = Run(args, Env{
		Getenv: func(string) string { return "" },
		Stdout: &out,
		Stderr: &errb,
		Width:  func() int { return width },
	})
	return code, out.String(), errb.String()
}

// exampleProject makes the project of the examples of docs/reference/cli.md.
func exampleProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []string{"src/auth", "src/db", "data"} {
		if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(d)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"src/main.go", "src/auth/session.ts", "src/db/conn.go", "data/2026.csv"} {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(f)), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	steps := [][]string{
		{"init"},
		{"set", "src/", "アプリケーションのソースコード"},
		{"set", "src/auth/", "認証(ログイン、セッション、トークン更新)", "--detail", "入口は index.ts。"},
		{"set", "src/auth/session.ts", "セッションの保存と有効期限の管理", "--detail",
			"セッションは Redis に保存し、アクセスのたびに有効期限を延長する。\n有効期限は config/auth.yaml の session.ttl で設定する。"},
		{"set", "src/auth/*_test.go", "同名ファイルのテスト"},
		{"set", "data/", "データ格納場所"},
		{"set", "data/*.csv", "CSVデータ"},
		{"set", "src/main.go", "エントリポイント"},
	}
	for _, s := range steps {
		if code, _, errs := runText(t, 0, append(s, "-C", dir)...); code != 0 {
			t.Fatalf("%q: %s", s, errs)
		}
	}
	return dir
}

func TestText(t *testing.T) {
	dir := exampleProject(t)
	tests := []struct {
		name string
		args []string
		code int
		want string
	}{
		{"ls (as in cli.md)", []string{"ls", "-C", dir, "src"}, 0, "" +
			"auth/     認証(ログイン、セッション、トークン更新)\n" +
			"db/       ! no description\n" +
			"main.go   エントリポイント\n"},
		{"tree with the mark", []string{"tree", "-C", filepath.Join(dir, "src", "auth")}, 0, "" +
			"  src/             アプリケーションのソースコード\n" +
			">   auth/          認証(ログイン、セッション、トークン更新)\n" +
			"      session.ts   セッションの保存と有効期限の管理\n" +
			"      *_test.go    同名ファイルのテスト\n" +
			"    main.go        エントリポイント\n" +
			"  data/            データ格納場所\n" +
			"    *.csv          CSVデータ\n"},
		{"tree with the mark on an ancestor", []string{"tree", "-C", filepath.Join(dir, "src", "db"), "src/", "-d", "2"}, 0, "" +
			"> src/        アプリケーションのソースコード  (here: src/db/)\n" +
			"    auth/     認証(ログイン、セッション、トークン更新)\n" +
			"    main.go   エントリポイント\n"},
		{"find (gap of 3)", []string{"find", "-C", dir, "セッション"}, 0, "" +
			"src/auth/             認証(ログイン、セッション、トークン更新)\n" +
			"src/auth/session.ts   セッションの保存と有効期限の管理\n"},
		{"find with more", []string{"find", "-C", dir, "セッション", "--limit", "1"}, 0, "" +
			"src/auth/   認証(ログイン、セッション、トークン更新)\n" +
			"… 1 more (use --limit)\n"},
		{"find nothing", []string{"find", "-C", dir, "nothing"}, 0, ""},
		{"show by a pattern (as in cli.md)", []string{"show", "-C", dir, "data/2026.csv"}, 0, "" +
			"data/2026.csv\n" +
			"  entry   data/*.csv\n" +
			"  desc    CSVデータ\n"},
		{"show with a detail (as in cli.md)", []string{"show", "-C", dir, "src/auth/session.ts"}, 0, "" +
			"src/auth/session.ts\n" +
			"  desc    セッションの保存と有効期限の管理\n" +
			"  detail  セッションは Redis に保存し、アクセスのたびに有効期限を延長する。\n" +
			"          有効期限は config/auth.yaml の session.ttl で設定する。\n"},
		{"show without a description", []string{"show", "-C", dir, "src/db"}, 0, "" +
			"src/db/\n" +
			"  desc    ! no description\n"},
		{"status", []string{"status", "-C", dir}, 1, "" +
			"untracked   dir   src/db/\n" +
			"untracked   file  src/db/conn.go\n" +
			"orphan      src/auth/*_test.go\n" +
			"\n" +
			"coverage    all: no (2)   dirs: no (1)\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, out, errs := runText(t, 0, tt.args...)
			if code != tt.code || out != tt.want || errs != "" {
				t.Errorf("%d %q\ngot:\n%s\nwant:\n%s", code, errs, out, tt.want)
			}
		})
	}
}

func TestTextShowLink(t *testing.T) {
	dir := exampleProject(t)
	if err := os.Symlink("releases/v2", filepath.Join(dir, "current")); err != nil {
		t.Skipf("cannot make a symbolic link here: %v", err)
	}
	want := "current\n  link    releases/v2\n  desc    ! no description\n"
	code, out, errs := runText(t, 0, "show", "-C", dir, "current")
	if code != 0 || out != want || errs != "" {
		t.Errorf("%d %q\ngot:\n%s\nwant:\n%s", code, errs, out, want)
	}
}

func TestTextShowEmptyDetailLine(t *testing.T) {
	dir := exampleProject(t)
	if code, _, errs := runText(t, 0, "set", "-C", dir, "src/main.go", "Entry", "--detail", "a\n\nb"); code != 0 {
		t.Fatal(errs)
	}
	// The empty line of the detail has no spaces at the end.
	want := "src/main.go\n  desc    Entry\n  detail  a\n\n          b\n"
	code, out, errs := runText(t, 0, "show", "-C", dir, "src/main.go")
	if code != 0 || out != want || errs != "" {
		t.Errorf("%d %q\ngot:\n%q\nwant:\n%q", code, errs, out, want)
	}
}

func TestTextStatus(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, errs := runText(t, 0, "init", "-C", dir); code != 0 {
		t.Fatal(errs)
	}
	// Nothing to report: the coverage line alone.
	if code, _, errs := runText(t, 0, "set", "-C", dir, "a.txt", "A"); code != 0 {
		t.Fatal(errs)
	}
	code, out, _ := runText(t, 0, "status", "-C", dir)
	if want := "coverage    all: yes   dirs: yes\n"; code != 0 || out != want {
		t.Errorf("%d\ngot:\n%s\nwant:\n%s", code, out, want)
	}

	// A desc 300 columns wide, written in the file by hand.
	desc := strings.Repeat("x", 300)
	if err := os.WriteFile(filepath.Join(dir, ".crube", "structure.txt"), []byte("a.txt\t"+desc+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, out, _ = runText(t, 0, "status", "-C", dir)
	if want := "invalid     a.txt  desc 300/256\n\ncoverage    all: yes   dirs: yes\n"; code != 1 || out != want {
		t.Errorf("%d\ngot:\n%s\nwant:\n%s", code, out, want)
	}
}

func TestCut(t *testing.T) {
	// Expected values from Python's unicodedata (east_asian_width W/F = 2,
	// categories Mn/Me/Cf = 0, else 1).
	tests := []struct {
		s     string
		avail int
		want  string
	}{
		{"abcdef", 4, "abc…"},
		{"abc", 3, "abc"},
		{"日本語", 4, "日…"},
		{"日本語", 5, "日本…"},
		{"abc", 0, "…"},
		// e + U+0301 (combining acute accent, 0 wide) + x + y
		{"e\xcc\x81xy", 2, "e\xcc\x81…"},
	}
	for _, tt := range tests {
		if got := cut(tt.s, tt.avail); got != tt.want {
			t.Errorf("cut(%q, %d) = %q, want %q", tt.s, tt.avail, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	dir := t.TempDir()
	for _, d := range []string{"src/auth", "src/db", "data"} {
		if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(d)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	steps := [][]string{
		{"init"},
		{"set", "src/", "アプリケーションのソースコード"},
		{"set", "src/auth/", "認証(ログイン、セッション、トークン更新)"},
		{"set", "src/auth/session.ts", "セッションの保存と有効期限の管理"},
		{"set", "data/", "データ格納場所"},
	}
	for _, s := range steps {
		if code, _, errs := runText(t, 0, append(s, "-C", dir)...); code != 0 {
			t.Fatalf("%q: %s", s, errs)
		}
	}

	// Expected lines from Python (a terminal 40 columns wide): the desc is
	// cut, "(here: …)" is kept.
	want := "" +
		"> src/             ア…  (here: src/db/)\n" +
		"    auth/          認証(ログイン、セッ…\n" +
		"      session.ts   セッションの保存と有…\n" +
		"  data/            データ格納場所\n"
	code, out, errs := runText(t, 40, "tree", "-C", filepath.Join(dir, "src", "db"))
	if code != 0 || out != want || errs != "" {
		t.Errorf("tree: %d %q\ngot:\n%s\nwant:\n%s", code, errs, out, want)
	}

	// At 41 columns the line with the note is exactly as wide as the terminal;
	// counting the 2 spaces before the note wrong would cut "アプ…" to "ア…".
	want = "" +
		"> src/             アプ…  (here: src/db/)\n" +
		"    auth/          認証(ログイン、セッシ…\n" +
		"      session.ts   セッションの保存と有…\n" +
		"  data/            データ格納場所\n"
	code, out, errs = runText(t, 41, "tree", "-C", filepath.Join(dir, "src", "db"))
	if code != 0 || out != want || errs != "" {
		t.Errorf("tree at 41: %d %q\ngot:\n%s\nwant:\n%s", code, errs, out, want)
	}

	// ls cuts on a terminal, but not with --full.
	code, out, _ = runText(t, 20, "ls", "-C", dir, "src")
	if want := "auth/   認証(ログイ…\ndb/     ! no descri…\n"; code != 0 || out != want {
		t.Errorf("ls: got:\n%s\nwant:\n%s", out, want)
	}
	code, out, _ = runText(t, 20, "ls", "--full", "-C", dir, "src")
	if want := "auth/   認証(ログイン、セッション、トークン更新)\ndb/     ! no description\n"; code != 0 || out != want {
		t.Errorf("ls --full: got:\n%s\nwant:\n%s", out, want)
	}

	// find cuts too.
	code, out, _ = runText(t, 20, "find", "-C", dir, "データ")
	if want := "data/   データ格納…\n"; code != 0 || out != want {
		t.Errorf("find: got:\n%s\nwant:\n%s", out, want)
	}
}
