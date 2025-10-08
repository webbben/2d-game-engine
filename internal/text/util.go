package text

import (
	"fmt"
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
			currentLine.Reset()
		}
	}

	return lines
}

// returns the height of a body of text, when split according to the given lineWidth
func GetStringLinesHeight(s string, f font.Face, lineWidthPx int) int {
	lines := ConvertStringToLines(s, f, lineWidthPx)
	height := 0

	for _, line := range lines {
		_, dy, _ := GetStringSize(line, f)
		height += dy
	}

	return height
}

func GetStringSize(s string, f font.Face) (dx int, dy int, baseline int) {
	bounds, advance := font.BoundString(f, s)
	return advance.Round(), (bounds.Max.Y - bounds.Min.Y).Round(), -bounds.Min.Y.Round()
}

// get size metrics for a font. here is what each value means:
//
// ascent: how far above baseline the font can go. basically, the "regular height" of the font, excluding any "descenders".
//
// descent: how far below baseline the font can go. basically, how far portions of letters like "y", "p", etc can go below baseline.
//
// height: the entire possible height - ascent and descent combined
func GetFontMetrics(f font.Face) (ascent int, descent int, height int) {
	metrics := f.Metrics()
	ascent = metrics.Ascent.Round()
	descent = metrics.Descent.Round()
	height = metrics.Height.Round()
	return ascent, descent, height
}

// since fonts can have lots of oddly sized symbols, this just gets a realistic expectation of the font's size.
//
// height: the height of a standard capital letter
//
// descent: how much common "descenders" ("y", "g", "p", etc) tends to go below baseline
func GetRealisticFontMetrics(f font.Face) (height int, descent int) {
	heightLetters := []string{"A", "B", "C", "D", "X", "Y", "Z"}
	// for height, get an average
	for _, l := range heightLetters {
		_, dy, _ := GetStringSize(l, f)
		height += dy
	}
	height = height / len(heightLetters)

	descentLetters := []string{"q", "g", "y", "p", "j"}
	// for descent, get max because we dont want any of these to get clipped on accident
	for _, l := range descentLetters {
		_, dy, _ := GetStringSize(fmt.Sprintf("%sX", l), f)
		dsc := dy - height
		if dsc > descent {
			descent = dsc
		}
	}

	return height, descent
}
