package text

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

type LineWriterStatus string

const (
	AwaitText  LineWriterStatus = "no_text_set"
	Writing    LineWriterStatus = "writing"
	AwaitPager LineWriterStatus = "awaiting_pager"
	TextDone   LineWriterStatus = "text_done"
)

// LineWriter is a tool to write lines of text that handle various functions like wrapping lines, etc.
type LineWriter struct {
	init            bool        // flag to indicate if this LineWriter was properly initialized
	sourceText      string      // full text the line writer is currently aiming to write
	textUpdateTimer int         // number of ticks until the next text character is added
	maxLineWidth    int         // max width (in pixels) of a line. lines will be
	maxHeight       int         // max height that the lineWriter's body of text can go. If the limit is met, lineWriter will split text into pages
	lineHeight      int         // the (max) height of a single line in this set of lines
	pageLineCount   int         // based on max height and line height, the number of lines that can fit in a single page
	fontFace        font.Face   // font to use when writing
	fgColor         color.Color // color of the text (foreground). defaults to black
	bgColor         color.Color // color of the text shadow. defaults to a semi-transparent gray.
	shadow          bool        // if set, text is drawn with the shadow (bgColor) effect

	linesToWrite      []string         // source text broken down into their lines
	pages             [][]string       // pages to write (linesToWrite broken down)
	currentPage       int              // the current page the lineWriter is writing
	currentLineNumber int              // the current line (of linesToWrite) that we are writing
	currentLineIndex  int              // the index of the current line we are writing
	writtenLines      []string         // the "output" that is actually drawn
	WritingStatus     LineWriterStatus // the current status of the lineWriter, regarding the text it is writing
	writeImmediately  bool
}

func (lw LineWriter) IsWriting() bool {
	return lw.WritingStatus == Writing
}

// CurrentDimensions gives the dimensions that the linewriter currently has - based on how much text has been written so far.
// If the linewriter is still writing, this may continue to change.
func (lw LineWriter) CurrentDimensions() (dx, dy int) {
	if lw.lineHeight == 0 {
		panic("tried to get dimensions of linewriter, but line height is 0")
	}
	if lw.maxLineWidth == 0 {
		panic("tried to get dimensions, but maxLineWidth was 0")
	}
	dy = len(lw.writtenLines) * lw.lineHeight
	return lw.maxLineWidth, dy
}

// NewLineWriter creates a new LineWriter.
// fg and bg colors can be left nil, in which case they assume the normal defaults (fg = black, bg = gray).
// set useShadow to true if you want the shadow effect to be used when drawing text.
func NewLineWriter(lineWidthPx, maxHeightPx int, f font.Face, fg, bg color.Color, useShadow bool, writeImmediately bool) LineWriter {
	minWidth, minHeight, _ := GetStringSize("ABC!/|", f)
	if lineWidthPx < minWidth {
		panic(fmt.Sprintf("lineWriter lineWidthPx must not be too small to draw text. minWidth determined by font: %v px", minWidth))
	}
	if maxHeightPx < minHeight {
		panic(fmt.Sprintf("lineWriter maxHeightPx must be be able to fit a single line of text. minHeight determined by font: %v px", minHeight))
	}

	lw := LineWriter{
		WritingStatus:    AwaitText,
		init:             true,
		maxLineWidth:     lineWidthPx,
		maxHeight:        maxHeightPx,
		fontFace:         f,
		fgColor:          fg,
		bgColor:          bg,
		shadow:           useShadow,
		writeImmediately: writeImmediately,
	}

	// prepare lines to write
	lw.linesToWrite = ConvertStringToLines(lw.sourceText, lw.fontFace, lw.maxLineWidth)
	lw.currentLineIndex = 0
	lw.currentLineNumber = 0
	lw.writtenLines = []string{""}

	// determine line height
	for _, line := range lw.linesToWrite {
		_, lineHeight, _ := GetStringSize(line, lw.fontFace)
		if lineHeight > lw.lineHeight {
			lw.lineHeight = lineHeight
		}
	}

	return lw
}

func (lw *LineWriter) SetSourceText(textToWrite string) {
	if textToWrite == "" {
		panic("tried to set empty text to lineWriter. to clear the lineWriter, use lw.Clear() instead.")
	}
	if lw.WritingStatus != AwaitText {
		panic("tried to set lineWriter text while it was in an invalid status. be sure to properly clear the lineWriter first with lw.Clear().")
	}

	textToWrite = strings.TrimSpace(textToWrite)

	lw.sourceText = textToWrite

	// prepare lines to write
	lw.linesToWrite = ConvertStringToLines(lw.sourceText, lw.fontFace, lw.maxLineWidth)
	lw.currentLineIndex = 0
	lw.currentLineNumber = 0
	lw.writtenLines = []string{""}

	// determine line height
	for _, line := range lw.linesToWrite {
		_, lineHeight, _ := GetStringSize(line, lw.fontFace)
		if lineHeight > lw.lineHeight {
			lw.lineHeight = lineHeight
		}
	}

	lw.pageLineCount = lw.maxHeight / lw.lineHeight

	// split lines into pages
	lw.pages = make([][]string, 0)
	page := []string{}
	lineCount := 0
	for _, line := range lw.linesToWrite {
		page = append(page, line)
		lineCount++
		if lineCount == lw.pageLineCount {
			lineCount = 0
			lw.pages = append(lw.pages, page)
			page = []string{}
		}
	}
	if len(page) > 0 {
		lw.pages = append(lw.pages, page)
	}

	lw.WritingStatus = Writing
}

// Clear fully clears source text and all written text
func (lw *LineWriter) Clear() {
	// unset all the values that are calculated when setting source text
	lw.linesToWrite = []string{}
	lw.currentLineIndex = 0
	lw.currentLineNumber = 0
	lw.writtenLines = []string{}
	lw.lineHeight = 0
	lw.currentPage = 0
	lw.pages = make([][]string, 0)
	lw.pageLineCount = 0

	// set status flag to indicate waiting for new text
	lw.WritingStatus = AwaitText
}

// Draw returns the last written Y position (for reference by other drawing functions)
func (lw LineWriter) Draw(screen *ebiten.Image, startX, startY int) int {
	y := startY
	for _, line := range lw.writtenLines {
		gray := color.RGBA{20, 20, 20, 75}
		y = y + lw.lineHeight
		DrawShadowText(screen, line, lw.fontFace, startX, y, color.Black, gray, -2, -2)
	}
	return y
}

func (lw *LineWriter) Update() {
	if lw.WritingStatus == AwaitText {
		return
	}
	if lw.WritingStatus == TextDone {
		return
	}
	if lw.WritingStatus == AwaitPager {
		return
	}

	if lw.WritingStatus == Writing {
		if lw.sourceText == "" {
			panic("lineWriter is writing but there is no source text set!")
		}
		if len(lw.pages) == 0 {
			panic("no pages for lineWriter")
		}
		if lw.currentPage > len(lw.pages)-1 {
			panic("current page index is too big!")
		}
		if lw.writeImmediately {
			lw.FastForward()
		}
		currentPage := lw.pages[lw.currentPage]
		if lw.currentLineNumber < len(currentPage) {
			if lw.currentLineIndex < len(currentPage[lw.currentLineNumber]) {
				lw.textUpdateTimer++
				if lw.textUpdateTimer >= 0 {
					// continue to write the current line
					nextChar := currentPage[lw.currentLineNumber][lw.currentLineIndex]
					lw.writtenLines[lw.currentLineNumber] += string(nextChar)
					lw.currentLineIndex++
					lw.textUpdateTimer = 0
				}
			} else {
				// go to the next line
				lw.currentLineNumber++
				lw.currentLineIndex = 0
				lw.writtenLines = append(lw.writtenLines, "") // start next output line
			}
		} else {
			// everything has been written for the current page
			if lw.currentPage < len(lw.pages)-1 {
				// there are more pages; wait for pager signal
				lw.WritingStatus = AwaitPager
				return
			}
			// no more pages; we're all done
			lw.WritingStatus = TextDone
		}
	}
}

func (lw *LineWriter) NextPage() {
	if lw.WritingStatus != AwaitPager {
		panic("tried to page lineWriter that isn't waiting for pager")
	}
	if lw.currentPage >= len(lw.pages)-1 {
		panic("tried to page lineWriter past max page number")
	}

	// clear written text (without deleting source text data)
	lw.currentLineIndex = 0
	lw.currentLineNumber = 0
	lw.writtenLines = []string{""}

	lw.currentPage++
	lw.WritingStatus = Writing
}

// FastForward finishes the current page
func (lw *LineWriter) FastForward() {
	currentPage := lw.pages[lw.currentPage]
	lw.writtenLines = []string{}
	lw.writtenLines = append(lw.writtenLines, currentPage...)

	lw.currentLineNumber = len(currentPage)
	lw.currentLineIndex = 0
}
