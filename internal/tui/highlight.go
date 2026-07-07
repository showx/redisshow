package tui

import (
	"strings"

	"github.com/rivo/tview"
)

func extractPatternLiterals(pattern string) []string {
	if pattern == "" || pattern == "*" {
		return nil
	}
	parts := strings.Split(pattern, "*")
	var literals []string
	for _, part := range parts {
		if part != "" {
			literals = append(literals, part)
		}
	}
	return literals
}

func highlightKeyName(key, pattern string) string {
	literals := extractPatternLiterals(pattern)
	if len(literals) == 0 {
		return tview.Escape(key)
	}

	marked := make([]bool, len(key))
	for _, literal := range literals {
		for i := 0; i+len(literal) <= len(key); i++ {
			if key[i:i+len(literal)] == literal {
				for j := i; j < i+len(literal); j++ {
					marked[j] = true
				}
			}
		}
	}

	var b strings.Builder
	inHighlight := false
	for i := 0; i < len(key); i++ {
		if marked[i] && !inHighlight {
			b.WriteString("[yellow]")
			inHighlight = true
		} else if !marked[i] && inHighlight {
			b.WriteString("[-]")
			inHighlight = false
		}
		b.WriteString(tview.Escape(string(key[i])))
	}
	if inHighlight {
		b.WriteString("[-]")
	}
	return b.String()
}
