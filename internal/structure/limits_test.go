package structure

import (
	"errors"
	"strings"
	"testing"
)

func TestCheckDesc(t *testing.T) {
	tests := []struct {
		name  string
		desc  string
		width int // 0: no error expected
	}{
		{"empty", "", 0},
		{"256 ASCII", strings.Repeat("a", 256), 0},
		{"257 ASCII", strings.Repeat("a", 257), 257},
		{"128 wide", strings.Repeat("あ", 128), 0},
		{"129 wide", strings.Repeat("あ", 129), 258},
		{"255 ASCII and one wide", strings.Repeat("a", 255) + "あ", 257},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckDesc(tt.desc)
			if tt.width == 0 {
				if err != nil {
					t.Fatalf("CheckDesc: %v", err)
				}
				return
			}
			var le *LimitError
			if !errors.As(err, &le) {
				t.Fatalf("CheckDesc = %v, want a *LimitError", err)
			}
			if *le != (LimitError{Field: "desc", Width: tt.width, Limit: 256}) {
				t.Errorf("got %+v, want desc width %d limit 256", *le, tt.width)
			}
		})
	}
}

func TestCheckDescNewline(t *testing.T) {
	for _, desc := range []string{"a\nb", "a\r\nb", "a\rb", "a\n"} {
		if err := CheckDesc(desc); !errors.Is(err, ErrDescNewline) {
			t.Errorf("CheckDesc(%q) = %v, want ErrDescNewline", desc, err)
		}
	}
}

func TestCheckDetail(t *testing.T) {
	tests := []struct {
		name   string
		detail string
		width  int // 0: no error expected
	}{
		{"empty", "", 0},
		{"4096 on one line", strings.Repeat("a", 4096), 0},
		{"4097 on one line", strings.Repeat("a", 4097), 4097},
		{"4096 over two lines", strings.Repeat("a", 4000) + "\n" + strings.Repeat("a", 96), 0},
		{"4097 over two lines", strings.Repeat("a", 4000) + "\n" + strings.Repeat("a", 97), 4097},
		{"line breaks are not counted", strings.Repeat("aaaa\n", 1024), 0},
		{"2048 wide", strings.Repeat("あ", 2048), 0},
		{"2049 wide", strings.Repeat("あ", 2049), 4098},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckDetail(tt.detail)
			if tt.width == 0 {
				if err != nil {
					t.Fatalf("CheckDetail: %v", err)
				}
				return
			}
			var le *LimitError
			if !errors.As(err, &le) {
				t.Fatalf("CheckDetail = %v, want a *LimitError", err)
			}
			if *le != (LimitError{Field: "detail", Width: tt.width, Limit: 4096}) {
				t.Errorf("got %+v, want detail width %d limit 4096", *le, tt.width)
			}
		})
	}
}

func TestLimitErrorMessage(t *testing.T) {
	// The wording approved for docs/reference/cli.md.
	err := &LimitError{Field: "desc", Width: 300, Limit: 256}
	if got, want := err.Error(), "desc is 300 columns wide (limit 256)"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCheckName(t *testing.T) {
	for _, name := range []string{"a", "a/", "*.csv", "*/", "[id].tsx", "my  file.txt", "|x", "x|", ".github/", "日本語.txt"} {
		if err := CheckName(name); err != nil {
			t.Errorf("CheckName(%q) = %v", name, err)
		}
	}
	for _, name := range []string{"", "/", ".", "./", "..", "../", "a/b", "a//", "a\tb", "a\nb", "a\rb", " a", "a ", " a/", "a /", "|", "|/", "| x"} {
		if err := CheckName(name); err == nil {
			t.Errorf("CheckName(%q) succeeded", name)
		}
	}
	// What CheckName accepts must read back as the same name.
	for _, name := range []string{"a", "a/", "*.csv", "[id].tsx", "my  file.txt", "|x"} {
		tree, err := Parse([]byte(name + "\tdesc\n"))
		if err != nil || len(tree.Entries) != 1 || tree.Entries[0].Name != name {
			t.Errorf("%q does not read back: %v", name, err)
		}
	}
}
