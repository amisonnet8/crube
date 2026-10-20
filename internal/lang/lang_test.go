package lang

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileName(t *testing.T) {
	tests := []struct{ tag, want string }{
		{"", "structure.txt"},
		{"en", "structure.txt"},
		{"ja", "structure_ja.txt"},
		{"zh-CN", "structure_zh-CN.txt"},
		{"ja-JP", "structure_ja-JP.txt"},
		{"EN", "structure_EN.txt"}, // no change of case
		{"en-US", "structure_en-US.txt"},
		{"sr-Latn-RS", "structure_sr-Latn-RS.txt"},
	}
	for _, tt := range tests {
		got, err := FileName(tt.tag)
		if err != nil || got != tt.want {
			t.Errorf("FileName(%q) = %q, %v; want %q", tt.tag, got, err, tt.want)
		}
	}
}

func TestInvalidTag(t *testing.T) {
	for _, tag := range []string{"../x", "ja/x", `ja\x`, "ja_JP", "ja ", " ja", "-ja", "ja-", "ja--JP", "abcdefghi", "日本", "ja.txt"} {
		if _, err := FileName(tag); err == nil {
			t.Errorf("FileName(%q) succeeded", tag)
		}
		if _, err := Select(t.TempDir(), tag); err == nil {
			t.Errorf("Select(%q) succeeded", tag)
		}
	}
	_, err := FileName("../x")
	if got, want := err.Error(), `invalid CRUBE_LANG "../x" (use a language tag such as ja or zh-CN)`; got != want {
		t.Errorf("message %q, want %q", got, want)
	}
}

func TestSelect(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"structure.txt", "structure_ja.txt", "structure_en.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "structure_fr.txt"), 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct{ tag, want string }{
		{"", "structure.txt"},
		{"en", "structure.txt"}, // structure_en.txt is not used for en
		{"ja", "structure_ja.txt"},
		{"de", "structure.txt"},    // no file for the language
		{"ja-JP", "structure.txt"}, // no fallback to ja
		{"fr", "structure.txt"},    // a directory is not a structure file
	}
	for _, tt := range tests {
		got, err := Select(dir, tt.tag)
		if err != nil || got != tt.want {
			t.Errorf("Select(%q) = %q, %v; want %q", tt.tag, got, err, tt.want)
		}
	}
}

func TestSelectWithoutAnyFile(t *testing.T) {
	// With no file at all, the default is chosen; reading it then reports
	// that it is missing.
	got, err := Select(t.TempDir(), "ja")
	if err != nil || got != DefaultFile {
		t.Errorf("Select = %q, %v; want %q", got, err, DefaultFile)
	}
}
