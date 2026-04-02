package styles

import (
	"strings"
	"unicode/utf8"
)

const MinColWidth = 22

// ComputeColumnWidths répartit la largeur du terminal entre nbCols colonnes.
// Retourne nil si le terminal est trop étroit.
func ComputeColumnWidths(termWidth, nbCols int) []int {
	if nbCols == 0 {
		return nil
	}
	// 2 marges globales + séparateurs entre colonnes
	available := termWidth - 2 - (nbCols - 1)
	if available < nbCols*MinColWidth {
		available = nbCols * MinColWidth
	}

	base := available / nbCols
	remainder := available % nbCols

	widths := make([]int, nbCols)
	for i := range widths {
		widths[i] = base
		if i < remainder {
			widths[i]++
		}
	}
	return widths
}

// TruncateTitle coupe le titre pour qu'il tienne dans width caractères.
func TruncateTitle(title string, width int) string {
	maxLen := width - 2
	if maxLen < 1 {
		return ""
	}
	if utf8.RuneCountInString(title) <= maxLen {
		return title
	}
	runes := []rune(title)
	return string(runes[:maxLen-1]) + "…"
}

// WrapText découpe le texte en lignes de longueur maximale width en coupant aux espaces.
func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var current []rune

	for _, word := range words {
		wr := []rune(word)
		// Mot trop long : le forcer sur une ou plusieurs lignes
		for len(wr) > width {
			if len(current) > 0 {
				lines = append(lines, string(current))
				current = nil
			}
			lines = append(lines, string(wr[:width]))
			wr = wr[width:]
		}
		if len(wr) == 0 {
			continue
		}
		if len(current) == 0 {
			current = wr
		} else if len(current)+1+len(wr) <= width {
			current = append(current, ' ')
			current = append(current, wr...)
		} else {
			lines = append(lines, string(current))
			current = wr
		}
	}
	if len(current) > 0 {
		lines = append(lines, string(current))
	}
	return strings.Join(lines, "\n")
}

// PadRight complète une chaîne avec des espaces jusqu'à width.
func PadRight(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}
