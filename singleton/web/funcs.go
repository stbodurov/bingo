package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"

	"github.com/stbodurov/bingo/singleton/internal/app"
	"github.com/stbodurov/bingo/singleton/internal/proof"
)

func funcs() template.FuncMap {
	return template.FuncMap{
		"num":   fmtNum,
		"pct":   func(f float64) string { return fmt.Sprintf("%.1f%%", f*100) },
		"money": func(f float64) string { return "$" + fmtMoney(f) },
		"ago":   ago,
		"date": func(s string) string {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return s
			}
			return t.Format("2 Jan 2006, 15:04")
		},
		"day": func(s string) string {
			t, err := time.Parse("2006-01-02", s)
			if err != nil {
				return s
			}
			return t.Format("Jan 2")
		},
		"short": func(s string, n int) string {
			if len(s) <= n {
				return s
			}
			return s[:n/2+2] + "…" + s[len(s)-n/2+1:]
		},
		"fp": func(pub string) string {
			e, err := proof.DecodeElementHex(pub)
			if err != nil {
				return pub
			}
			return proof.Fingerprint(e)
		},
		"pseudonym": func(n string) string {
			if len(n) < 12 {
				return n
			}
			return n[:4] + "·" + n[4:8] + "·" + n[8:12]
		},
		"upper": strings.ToUpper,
		"title": func(s string) string {
			if s == "" {
				return s
			}
			return strings.ToUpper(s[:1]) + strings.ReplaceAll(s[1:], "_", " ")
		},
		"join": strings.Join,
		"json": func(v any) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
		"setLabel": func(name string) string {
			if d, ok := app.SetByName(name); ok {
				return d.Label
			}
			return name
		},
		"statusClass": func(s string) string {
			return map[string]string{"verified": "good", "duplicate": "warn", "failed": "bad", "expired": "muted", "pending": "info", "delivered": "good", "approved": "good", "rejected": "bad"}[s]
		},
		"add":  func(a, b int) int { return a + b },
		"mulf": func(a int, b float64) float64 { return float64(a) * b },
		"sub":  func(a, b int) int { return a - b },
		"div":  func(a, b int) float64 { if b == 0 { return 0 }; return float64(a) / float64(b) },
		"widthPct": func(v, max int) string {
			if max <= 0 {
				return "0"
			}
			return fmt.Sprintf("%.1f", math.Max(1.5, float64(v)/float64(max)*100))
		},
		"maxOf": func(m map[string]int) int {
			x := 0
			for _, v := range m {
				if v > x {
					x = v
				}
			}
			return x
		},
		"has": func(m map[string]bool, k string) bool { return m[k] },
		"dict": func(kv ...any) map[string]any {
			m := map[string]any{}
			for i := 0; i+1 < len(kv); i += 2 {
				m[fmt.Sprint(kv[i])] = kv[i+1]
			}
			return m
		},
		"stacked":   stackedBars,
		"sparkline": sparkline,
		"now":       func() string { return time.Now().UTC().Format(time.RFC3339) },
		"sliceOf":   func(s ...string) []string { return s },
		"isActive":  func(notAfter, revoked string) bool { return revoked == "" && notAfter > time.Now().UTC().Format(time.RFC3339) },
	}
}

func fmtNum(v any) string {
	var n int64
	switch x := v.(type) {
	case int:
		n = int64(x)
	case int64:
		n = x
	case float64:
		n = int64(x)
	default:
		return fmt.Sprint(v)
	}
	s := fmt.Sprint(n)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}

func fmtMoney(f float64) string {
	whole := int64(f)
	cents := int64(math.Round((f - float64(whole)) * 100))
	if cents == 100 {
		whole++
		cents = 0
	}
	return fmt.Sprintf("%s.%02d", fmtNum(whole), cents)
}

func ago(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func niceMax(v int) int {
	if v <= 5 {
		return 5
	}
	mag := math.Pow(10, math.Floor(math.Log10(float64(v))))
	for _, m := range []float64{1, 2, 2.5, 5, 10} {
		if float64(v) <= m*mag {
			return int(m * mag)
		}
	}
	return v
}

// stackedBars renders daily session outcomes as an accessible SVG column chart:
// ≤24px columns, 2px surface gaps between segments, 4px rounded top, hairline grid.
func stackedBars(days []app.DayStat) template.HTML {
	const w, h, padL, padB, padT = 760.0, 220.0, 40.0, 24.0, 10.0
	max := 0
	for _, d := range days {
		if t := d.Verified + d.Duplicate + d.Failed + d.Abandoned; t > max {
			max = t
		}
	}
	top := niceMax(max)
	n := float64(len(days))
	if n == 0 {
		return ""
	}
	band := (w - padL) / n
	bw := math.Min(24, band*0.62)
	y := func(v float64) float64 { return padT + (1-v/float64(top))*(h-padT-padB) }
	var b strings.Builder
	fmt.Fprintf(&b, `<svg class="chart" viewBox="0 0 %.0f %.0f" role="img" aria-label="Verification outcomes per day">`, w, h)
	for i := 0; i <= 4; i++ {
		v := float64(top) * float64(i) / 4
		fmt.Fprintf(&b, `<line class="grid" x1="%.0f" x2="%.0f" y1="%.1f" y2="%.1f"/><text class="tick" x="%.0f" y="%.1f" text-anchor="end">%s</text>`, padL, w, y(v), y(v), padL-8, y(v)+4, fmtNum(int(v)))
	}
	type seg struct {
		v     int
		cls   string
		label string
	}
	for i, d := range days {
		cx := padL + band*float64(i) + band/2
		segs := []seg{{d.Verified, "s-good", "verified"}, {d.Duplicate, "s-warn", "duplicate"}, {d.Failed, "s-bad", "failed"}, {d.Abandoned, "s-muted", "abandoned"}}
		base := 0.0
		lastIdx := -1
		for j, sg := range segs {
			if sg.v > 0 {
				lastIdx = j
			}
		}
		total := d.Verified + d.Duplicate + d.Failed + d.Abandoned
		tip := fmt.Sprintf("%s — %d verified · %d duplicate · %d failed · %d abandoned", d.Day, d.Verified, d.Duplicate, d.Failed, d.Abandoned)
		fmt.Fprintf(&b, `<g class="col" data-tip="%s"><title>%s</title><rect class="hit" x="%.1f" y="0" width="%.1f" height="%.0f"/>`, template.HTMLEscapeString(tip), template.HTMLEscapeString(tip), padL+band*float64(i), band, h-padB)
		for j, sg := range segs {
			if sg.v == 0 {
				continue
			}
			y0 := y(base)
			y1 := y(base + float64(sg.v))
			hh := y0 - y1
			if j != lastIdx {
				hh -= 2 // 2px surface gap above each non-top segment
			}
			if hh < 0.5 {
				hh = 0.5
			}
			x := cx - bw/2
			if j == lastIdx {
				r := math.Min(4, math.Min(bw/2, hh))
				fmt.Fprintf(&b, `<path class="%s" d="M%.1f,%.1fv%.1fq0,%.1f %.1f,%.1fh%.1fq%.1f,0 %.1f,%.1fv%.1fz"/>`, sg.cls, x, y0, -(hh - r), -r, r, -r, bw-2*r, r, r, r, hh-r)
			} else {
				fmt.Fprintf(&b, `<rect class="%s" x="%.1f" y="%.1f" width="%.1f" height="%.1f"/>`, sg.cls, x, y0-hh, bw, hh)
			}
			base += float64(sg.v)
		}
		b.WriteString(`</g>`)
		if i%5 == 0 || i == len(days)-1 {
			t, _ := time.Parse("2006-01-02", d.Day)
			fmt.Fprintf(&b, `<text class="tick" x="%.1f" y="%.0f" text-anchor="middle">%s</text>`, cx, h-6, t.Format("Jan 2"))
		}
		_ = total
	}
	fmt.Fprintf(&b, `<line class="axis" x1="%.0f" x2="%.0f" y1="%.1f" y2="%.1f"/></svg>`, padL, w, y(0), y(0))
	return template.HTML(b.String())
}

// sparkline renders verified counts per day as a small line.
func sparkline(days []app.DayStat) template.HTML {
	if len(days) < 2 {
		return ""
	}
	const w, h = 120.0, 28.0
	max := 1
	for _, d := range days {
		if d.Verified > max {
			max = d.Verified
		}
	}
	var pts []string
	for i, d := range days {
		x := 2 + float64(i)/float64(len(days)-1)*(w-4)
		y := h - 3 - float64(d.Verified)/float64(max)*(h-6)
		pts = append(pts, fmt.Sprintf("%.1f,%.1f", x, y))
	}
	last := pts[len(pts)-1]
	xy := strings.Split(last, ",")
	return template.HTML(fmt.Sprintf(`<svg class="spark" viewBox="0 0 %.0f %.0f" preserveAspectRatio="none" aria-hidden="true"><polyline points="%s"/><circle cx="%s" cy="%s" r="2.5"/></svg>`, w, h, strings.Join(pts, " "), xy[0], xy[1]))
}
