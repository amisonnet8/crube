package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amisonnet8/crube/internal/textwidth"
)

// demoSession is what docs/assets/demo.svg (the animated terminal at the top
// of README.md) shows, run in the project of cli.md: each scene is a screen,
// and each line of it a command typed in the same shell. demo_ja.svg, of
// README_ja.md, shows the same in the project described in Japanese.
func demoSession(l docLang) [][]string {
	return [][]string{
		{"cd src/auth", "crube tree"},
		{"crube find " + l.findWord},
		{"crube show session.ts"},
		{"crube status"},
	}
}

// TestDemoSVG keeps docs/assets/demo.svg and demo_ja.svg the same as what
// crube prints: it runs demoSession and draws it. With UPDATE_DEMO=1 (qsoku
// demo), it writes the files instead of comparing.
func TestDemoSVG(t *testing.T) {
	for _, l := range docLangs {
		name := "docs/assets/demo" + l.suffix + ".svg"
		t.Run(name, func(t *testing.T) {
			root, link := cliProjectIn(t, l)
			if !link {
				t.Skip("symbolic links cannot be made here, and the status of the demo shows one")
			}
			s := newShell(t, root)
			var scenes [][]demoStep
			for _, cmds := range demoSession(l) {
				var scene []demoStep
				for _, cmd := range cmds {
					scene = append(scene, demoStep{cmd: cmd, out: splitOutput(s.exec(cmd))})
				}
				scenes = append(scenes, scene)
			}
			got := drawDemo(scenes)

			path := filepath.Join("..", filepath.FromSlash(name))
			if os.Getenv("UPDATE_DEMO") == "1" {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(want) != got {
				t.Errorf("%s is not what crube prints now: run qsoku demo, and look at the new image before committing it", name)
			}
		})
	}
}

type demoStep struct {
	cmd string
	out []string
}

// The look of the terminal: sizes in pixels, times in seconds.
const (
	demoFont      = 14.0
	demoCol       = 8.4 // the advance of one column (0.6 em, that of common monospace fonts)
	demoLine      = 20.0
	demoPad       = 20.0
	demoBar       = 36.0
	demoMinWidth  = 640.0
	demoStart     = 0.6  // before the first scene
	demoTypeCol   = 0.05 // to type one column
	demoAfterType = 0.35 // from the end of typing to the output
	demoReadLine  = 0.18 // the time to read one line of output
	demoReadMin   = 1.4
	demoSceneEnd  = 0.6 // between the last output of a scene and the next scene
	demoLoopEnd   = 1.5 // after the last scene, before it all starts again
)

// demoAnim is one CSS animation: the element is hidden before from and after
// to (seconds), shown in between.
type demoAnim struct {
	class    string
	from, to float64
}

// drawDemo draws the scenes as an SVG of a terminal window that types each
// command, shows its output, and goes to the next scene, in a loop. Each
// character is put at its column (wide characters take two), so the columns
// line up whatever width the font gives to wide characters.
func drawDemo(scenes [][]demoStep) string {
	cols, rows := 0, 0
	for _, scene := range scenes {
		n := 0
		for _, st := range scene {
			cols = max(cols, textwidth.String("$ "+st.cmd))
			for _, l := range st.out {
				cols = max(cols, textwidth.String(l))
			}
			n += 1 + len(st.out)
		}
		rows = max(rows, n)
	}
	width := max(demoMinWidth, 2*demoPad+float64(cols)*demoCol)
	height := demoBar + 2*demoPad + float64(rows)*demoLine

	var body strings.Builder
	var anims []demoAnim
	typing := map[string][2]float64{} // the class of a cover: when typing starts and ends
	t := demoStart
	for i, scene := range scenes {
		sceneClass := fmt.Sprintf("s%d", i)
		sceneFrom := t
		fmt.Fprintf(&body, "<g class=\"%s\">\n", sceneClass)
		row := 0
		for j, st := range scene {
			id := fmt.Sprintf("%d-%d", i, j)
			y := demoBar + demoPad + float64(row)*demoLine
			typed := float64(textwidth.String(st.cmd)) * demoTypeCol

			// The command line appears with its prompt; a cover over the
			// command shrinks to the right, column by column, as if typed.
			anims = append(anims, demoAnim{"c" + id, t, -1})
			fmt.Fprintf(&body, "<g class=\"c%s\">\n", id)
			body.WriteString(demoText(y, []demoSeg{{"$", "prompt"}, {" " + st.cmd, "cmd"}}))
			fmt.Fprintf(&body, "<rect class=\"k%s cover\" x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\"/>\n</g>\n",
				id, demoPad+2*demoCol, y, float64(textwidth.String(st.cmd))*demoCol+2, demoLine)
			typing["k"+id] = [2]float64{t, t + typed}
			t += typed + demoAfterType
			row++

			if len(st.out) == 0 {
				continue
			}
			anims = append(anims, demoAnim{"o" + id, t, -1})
			fmt.Fprintf(&body, "<g class=\"o%s\">\n", id)
			for _, l := range st.out {
				body.WriteString(demoText(demoBar+demoPad+float64(row)*demoLine, demoColor(l)))
				row++
			}
			body.WriteString("</g>\n")
			t += max(demoReadMin, float64(len(st.out))*demoReadLine)
		}
		body.WriteString("</g>\n")
		anims = append(anims, demoAnim{sceneClass, sceneFrom, t + demoSceneEnd})
		t += demoSceneEnd
	}
	total := t + demoLoopEnd

	var css strings.Builder
	pct := func(s float64) string { return fmt.Sprintf("%.2f%%", s/total*100) }
	after := func(s float64) string { return fmt.Sprintf("%.2f%%", s/total*100+0.01) }
	for _, a := range anims {
		to := total
		if a.to >= 0 {
			to = a.to
		}
		fmt.Fprintf(&css, ".%s{animation:%s %.2fs linear infinite}\n", a.class, a.class, total)
		if to < total {
			fmt.Fprintf(&css, "@keyframes %s{0%%,%s{opacity:0}%s,%s{opacity:1}%s,100%%{opacity:0}}\n",
				a.class, pct(a.from), after(a.from), pct(to), after(to))
		} else {
			fmt.Fprintf(&css, "@keyframes %s{0%%,%s{opacity:0}%s,100%%{opacity:1}}\n", a.class, pct(a.from), after(a.from))
		}
	}
	for i, scene := range scenes {
		for j, st := range scene {
			k := fmt.Sprintf("k%d-%d", i, j)
			span := typing[k]
			n := max(1, textwidth.String(st.cmd))
			fmt.Fprintf(&css, ".%s{animation:%s %.2fs infinite}\n", k, k, total)
			fmt.Fprintf(&css, "@keyframes %s{0%%,%s{transform:scaleX(1);animation-timing-function:steps(%d,end)}%s,100%%{transform:scaleX(0)}}\n",
				k, pct(span[0]), n, pct(span[1]))
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%.0f\" height=\"%.0f\" viewBox=\"0 0 %.0f %.0f\" role=\"img\" aria-label=\"crube in a terminal: tree, find, show and status\">\n",
		width, height, width, height)
	b.WriteString("<!-- Made by TestDemoSVG (e2e/demo_test.go) from what crube prints; run qsoku demo to make it again. -->\n")
	b.WriteString("<style>\n")
	b.WriteString("text{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,\"DejaVu Sans Mono\",\"Noto Sans Mono CJK JP\",monospace;font-size:14px;fill:#c9d1d9}\n")
	b.WriteString(".prompt{fill:#3fb950;font-weight:bold}.cmd{fill:#f0f6fc;font-weight:bold}.mark{fill:#58a6ff;font-weight:bold}")
	b.WriteString(".warn{fill:#d29922}.err{fill:#f85149}.label{fill:#79c0ff}.dim{fill:#8b949e}\n")
	b.WriteString(".cover{fill:#0d1117;transform-box:fill-box;transform-origin:right}\n")
	b.WriteString(css.String())
	b.WriteString("@media (prefers-reduced-motion:reduce){*{animation:none!important}}\n")
	b.WriteString("</style>\n")
	fmt.Fprintf(&b, "<rect x=\"0.5\" y=\"0.5\" width=\"%.0f\" height=\"%.0f\" rx=\"10\" fill=\"#0d1117\" stroke=\"#30363d\"/>\n", width-1, height-1)
	fmt.Fprintf(&b, "<path d=\"M0.5 %.1fH%.1f\" stroke=\"#30363d\"/>\n", demoBar, width-0.5)
	b.WriteString("<circle cx=\"20\" cy=\"18\" r=\"6\" fill=\"#ff5f56\"/><circle cx=\"40\" cy=\"18\" r=\"6\" fill=\"#ffbd2e\"/><circle cx=\"60\" cy=\"18\" r=\"6\" fill=\"#27c93f\"/>\n")
	fmt.Fprintf(&b, "<text x=\"%.1f\" y=\"23\" text-anchor=\"middle\" class=\"dim\" style=\"font-size:12px\">crube</text>\n", width/2)
	b.WriteString(body.String())
	b.WriteString("</svg>\n")
	return b.String()
}

// demoSeg is a run of text drawn in one class.
type demoSeg struct {
	text, class string
}

// demoColor splits a line of output into the runs to color: the mark of the
// current place, the kinds of status, the labels of show, "! no description".
func demoColor(line string) []demoSeg {
	if rest, ok := strings.CutPrefix(line, ">"); ok {
		return append([]demoSeg{{">", "mark"}}, demoColor(rest)...)
	}
	for word, class := range map[string]string{"untracked": "warn", "orphan": "err", "invalid": "err", "coverage": "label"} {
		if rest, ok := strings.CutPrefix(line, word); ok {
			return []demoSeg{{word, class}, {rest, ""}}
		}
	}
	for _, label := range []string{"entry", "desc", "detail", "link"} {
		if strings.HasPrefix(line, "  "+label+" ") {
			return append([]demoSeg{{"  " + label, "label"}}, demoColor(line[len("  "+label):])...)
		}
	}
	if before, after, ok := strings.Cut(line, "! no description"); ok {
		return []demoSeg{{before, ""}, {"! no description", "warn"}, {after, ""}}
	}
	return []demoSeg{{line, ""}}
}

// demoText draws one line at the baseline of the row whose top is y. Spaces
// are not drawn: every other character gets its own x, from its column.
func demoText(y float64, segs []demoSeg) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<text y=\"%.1f\">", y+demoFont+1)
	col := 0
	for _, s := range segs {
		var xs []string
		var chars strings.Builder
		for _, r := range s.text {
			if r != ' ' {
				xs = append(xs, fmt.Sprintf("%.1f", demoPad+float64(col)*demoCol))
				chars.WriteString(demoEscape(r))
			}
			col += textwidth.Rune(r)
		}
		if len(xs) == 0 {
			continue
		}
		class := ""
		if s.class != "" {
			class = fmt.Sprintf(" class=\"%s\"", s.class)
		}
		fmt.Fprintf(&b, "<tspan x=\"%s\"%s>%s</tspan>", strings.Join(xs, " "), class, chars.String())
	}
	b.WriteString("</text>\n")
	return b.String()
}

func demoEscape(r rune) string {
	switch r {
	case '&':
		return "&amp;"
	case '<':
		return "&lt;"
	case '>':
		return "&gt;"
	}
	return string(r)
}
