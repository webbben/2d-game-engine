package text

import (
	"strings"

	"golang.org/x/image/font"
)

func ConvertStringToLines(s string, f font.Face, lineWidthPx int) []string {
	if lineWidthPx == 0 {
		panic("ConvertStringtoLines: lineWidthPx is 0!")
	}
	// get all the words (space separated) and find out how many words can be fit in a line
	words := strings.Fields(s)
	var currentLine strings.Builder
	currentLineWidth := 0
	lines := []string{}

	for i, w := range words {
		if i < len(words)-1 {
			w += " " // each word has a space
		}
		wordDx, _, _ := GetStringSize(w, f)
		if currentLineWidth+wordDx > lineWidthPx {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLineWidth = 0
		}
		currentLineWidth += wordDx
		currentLine.WriteString(w)
	}

	lines = append(lines, currentLine.String())

	return lines
}

func GetStringSize(s string, f font.Face) (dx int, dy int, baseline int) {
	bounds, advance := font.BoundString(f, s)
	return advance.Round(), (bounds.Max.Y - bounds.Min.Y).Round(), -bounds.Min.Y.Round()
}
