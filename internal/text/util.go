package text

import (
	"strings"

	"golang.org/x/image/font"
)

// splits a string into lines based on the size of the rendered text with the given font.
// also handles "\n" new line carriages as explicit line breaks.
func ConvertStringToLines(s string, f font.Face, lineWidthPx int) []string {
	if lineWidthPx == 0 {
		panic("ConvertStringtoLines: lineWidthPx is 0!")
	}
	// find all explicit line breaks first
	paragraphs := strings.Split(s, "\n")

	lines := []string{}
	var currentLine strings.Builder

	// within each explicit line-broken section, make sure those sections don't spill too long either
	for _, p := range paragraphs {
		if p == "" {
			// handle empty lines
			lines = append(lines, "")
			continue
		}

		// get all the words (space separated) and find out how many words can be fit in a line
		words := strings.Fields(p)
		currentLineWidth := 0

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

		if currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
		}
	}

	return lines
}

func GetStringSize(s string, f font.Face) (dx int, dy int, baseline int) {
	bounds, advance := font.BoundString(f, s)
	return advance.Round(), (bounds.Max.Y - bounds.Min.Y).Round(), -bounds.Min.Y.Round()
}
