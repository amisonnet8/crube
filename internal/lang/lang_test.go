package lang

import (
	"os"
	"path/filepath"
	"slices"
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

func TestPick(t *testing.T) {
	tests := []struct{ arg, name, env, want, msg string }{
		{"", "--lang", "", "", ""},
		{"ja", "--lang", "", "ja", ""},
		{"ja", "--lang", "fr", "ja", ""}, // the argument comes first
		{"", "--lang", "fr", "fr", ""},
		{"ja", "--lang", "x_y", "ja", ""}, // CRUBE_LANG is not used, so not checked
		{"ja_JP", "--lang", "", "", `invalid --lang "ja_JP" (use a language tag such as ja or zh-CN)`},
		{"ja_JP", "lang", "fr", "", `invalid lang "ja_JP" (use a language tag such as ja or zh-CN)`},
		{"", "--lang", "ja_JP", "", `invalid CRUBE_LANG "ja_JP" (use a language tag such as ja or zh-CN)`},
	}
	for _, tt := range tests {
		got, err := Pick(tt.arg, tt.name, tt.env)
		if tt.msg != "" {
			if err == nil || err.Error() != tt.msg {
				t.Errorf("Pick(%q, %q, %q): got %v, want %q", tt.arg, tt.name, tt.env, err, tt.msg)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Errorf("Pick(%q, %q, %q) = %q, %v; want %q", tt.arg, tt.name, tt.env, got, err, tt.want)
		}
	}
}

func TestTags(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"ignore", "notes_ja.txt", "notes.txt", "structure_ja.txt", "structure_zh-CN.txt", "structure_en.txt", "structure_ja_JP.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "structure_fr.txt"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Tags(dir)
	if err != nil || !slices.Equal(got, []string{"ja", "zh-CN"}) {
		t.Errorf("without structure.txt: %q, %v", got, err)
	}
	if err := os.WriteFile(filepath.Join(dir, DefaultFile), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err = Tags(dir)
	if err != nil || !slices.Equal(got, []string{"en", "ja", "zh-CN"}) {
		t.Errorf("with structure.txt: %q, %v", got, err)
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
