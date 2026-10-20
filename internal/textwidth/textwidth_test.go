package textwidth

import "testing"

// The expected widths were computed with Python's unicodedata (East Asian
// Width; W and F are 2, categories Mn, Me and Cf are 0, the rest 1), not with
// this package.
func TestString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"empty", "", 0},
		{"ASCII", "abc", 3},
		{"wide", "日本語", 6},
		{"halfwidth katakana", "ｱｲｳ", 3},
		{"fullwidth letter", "Ａ", 2},
		{"ambiguous counts as 1", "○①α", 3},
		{"precomposed e acute (U+00E9, ambiguous)", "\xc3\xa9", 1},
		{"e and combining acute (U+0301)", "e\xcc\x81", 1},
		{"zero width joiner (U+200D)", "a\xe2\x80\x8db", 2},
		{"emoji", "😀", 2},
		{"tab", "\t", 1},
		{"mixed", "認証(ログイン、セッション、トークン更新)", 40},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(tt.s); got != tt.want {
				t.Errorf("String(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestRune(t *testing.T) {
	for r, want := range map[rune]int{'a': 1, 'あ': 2, 'ｱ': 1, '○': 1, 0x0301: 0, 0x200d: 0} {
		if got := Rune(r); got != want {
			t.Errorf("Rune(%U) = %d, want %d", r, got, want)
		}
	}
}
