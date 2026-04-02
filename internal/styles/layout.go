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

// PadRight complète une chaîne avec des espaces jusqu'à width.
func PadRight(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}
