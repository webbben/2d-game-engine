package text

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"golang.org/x/image/font"
)

type LineWriterStatus string

const (
	AwaitText  LineWriterStatus = "no_text_set"
	Writing    LineWriterStatus = "writing"
	AwaitPager LineWriterStatus = "awaiting_pager"
	TextDone   LineWriterStatus = "text_done"
)

const ticksPerTextWrite = 1

var slowDownChars = map[rune]bool{
	'.': true,
	'!': true,
	'?': true,
	'-': true,
}

func isSlowdownChar(r rune) bool {
	_, exists := slowDownChars[r]
	return exists
}

// LineWriter is a tool to write lines of text that handle various functions like wrapping lines, etc.
type LineWriter struct {
	audioman         *audio.AudioManager
	textBlipSfx      defs.SoundID
	ticksTilBlip     int  // remaining ticks until next "blip" sfx
	blipTickInterval int  // number of ticks between blip sound while text is writing
	specialSymbols   bool // if true, lineWriter handles special symbols like underscores, square brackets, etc. otherwise, it just draws those symbols as is.

	init             bool        // flag to indicate if this LineWriter was properly initialized
	sourceText       string      // full text the line writer is currently aiming to write
	textUpdateTimer  int         // number of ticks until the next text character is added
	addedUpdateTimer int         // added ticks for next update, to slow down pace of writing
	maxLineWidth     int         // max width (in pixels) of a line. lines will be
	maxHeight        int         // max height that the lineWriter's body of text can go. If the limit is met, lineWriter will split text into pages
	lineHeight       int         // the (max) height of a single line in this set of lines
	pageLineCount    int         // based on max height and line height, the number of lines that can fit in a single page
	fontFace         font.Face   // font to use when writing
	fgColor          color.Color // color of the text (foreground). defaults to black
	currentFgColor   color.Color // color currently being used to write the next rune in Fg. this can change based on text notation for asides or links, etc.
	bgColor          color.Color // color of the text shadow. defaults to a semi-transparent gray.
	linkColor        color.Color // color of links in text
	shadow           bool        // if set, text is drawn with the shadow (bgColor) effect
	currentShadow    bool        // we might temporarily turn off shadow, when drawing de-emphasized text for example

	linesToWrite      []string         // source text broken down into their lines
	pages             [][]string       // pages to write (linesToWrite broken down)
	currentPage       int              // the current page the lineWriter is writing
	currentLineNumber int              // the current line (of linesToWrite) that we are writing
	currentLineIndex  int              // the index of the current line we are writing
	currentLineRunes  []rune           // the runes on the current line that is being written. for 'typewriter' style where one character is drawn at a time.
	writtenLines      []string         // the "output" that is actually drawn. Note that this isn't used to draw the actual text anymore; just keeping in case its useful for debugging.
	linkPositions     []LinkPos        // positions of all links drawn by the linkWriter.
	currentLinkPos    *LinkPos         // if currently drawing a link, this is created to record start X/Y.
	WritingStatus     LineWriterStatus // the current status of the lineWriter, regarding the text it is writing
	writeImmediately  bool

	textImg          *ebiten.Image // image where the text is actually drawn.
	cursorX, cursorY int           // where (on the text image) the cursor is that draws the next string or character
	cursorOffsetX    int           // how much to offset the starting cursor pos
}

func (lw LineWriter) IsWriting() bool {
	return lw.WritingStatus == Writing
}

// CurrentDimensions gives the dimensions that the linewriter currently has - based on how much text has been written so far.
// If the linewriter is still writing, this may continue to change.
func (lw LineWriter) CurrentDimensions() (dx, dy int) {
	if lw.lineHeight == 0 {
		panic("tried to get dimensions of linewriter, but line height is 0. did you set the source text yet?")
	}
	if lw.maxLineWidth == 0 {
		panic("tried to get dimensions, but maxLineWidth was 0")
	}
	if lw.cursorX == lw.cursorOffsetX && lw.cursorY != lw.lineHeight {
		// the cursor is sitting at the new line position, but the current line doesn't have any text
		return lw.maxLineWidth, lw.cursorY - lw.lineHeight
	}
	return lw.maxLineWidth, lw.cursorY
}

// NewLineWriter creates a new LineWriter.
// fg and bg colors can be left nil, in which case they assume the normal defaults (fg = black, bg = gray).
// set useShadow to true if you want the shadow effect to be used when drawing text.
//
// Audioman is optional; only needed if you set TextBlipSfx.
func NewLineWriter(audioman *audio.AudioManager, params config.LineWriterParams) LineWriter {
	// NOTE: audioman may be nil if linewriter is used by a place that doesn't expect to use sound effects. sound effects are mainly used by dialog.

	if params.FontFace == nil {
		logz.Panic("font face was nil!")
	}
	minWidth, minHeight, _ := GetStringSize("ABC!/|", params.FontFace)
	if params.LineWidthPx < minWidth {
		panic(fmt.Sprintf("lineWriter lineWidthPx must not be too small to draw text. minWidth determined by font: %v px", minWidth))
	}
	if params.MaxHeightPx < minHeight {
		panic(fmt.Sprintf("lineWriter maxHeightPx must be be able to fit a single line of text. minHeight determined by font: %v px", minHeight))
	}
	if params.FgColor == nil {
		params.FgColor = color.Black
	}
	if params.BgColor == nil {
		params.BgColor = color.RGBA{0, 0, 0, 70}
	}
	if params.LinkColor == nil {
		params.LinkColor = color.RGBA{0, 0, 255, 255}
	}

	lw := LineWriter{
		WritingStatus:    AwaitText,
		init:             true,
		maxLineWidth:     params.LineWidthPx,
		maxHeight:        params.MaxHeightPx,
		fontFace:         params.FontFace,
		fgColor:          params.FgColor,
		bgColor:          params.BgColor,
		linkColor:        params.LinkColor,
		shadow:           params.UseShadow,
		currentShadow:    params.UseShadow,
		writeImmediately: params.WriteImmediately,
		audioman:         audioman,
		specialSymbols:   params.SupportSpecialSymbols,
	}

	lw.textImg = ebiten.NewImage(lw.maxLineWidth, lw.maxHeight)
	lw.cursorOffsetX = 2

	if params.TextBlipSfx != "" {
		if params.TextBlipTickInterval < 0 {
			panic("tick interval was < 0")
		}
		if lw.audioman == nil {
			logz.Panic("audioman was nil. ensure it's passed into NewLineWriter if this linewriter is supposed to support SFX.")
		}
		lw.textBlipSfx = params.TextBlipSfx
		lw.blipTickInterval = params.TextBlipTickInterval
	}

	return lw
}

// SetSourceText sets the text for the lineWriter to write. It's expected that you will call Clear first if there is already written text.
// If you want to clear the lineWriter of all text, you should use that Clear function too - don't try passing an empty string in here.
func (lw *LineWriter) SetSourceText(textToWrite string) {
	if textToWrite == "" {
		panic("tried to set empty text to lineWriter. to clear the lineWriter, use lw.Clear() instead.")
	}
	if lw.WritingStatus != AwaitText {
		panic("tried to set lineWriter text while it was in an invalid status. be sure to properly clear the lineWriter first with lw.Clear().")
	}

	lw.linkPositions = []LinkPos{}

	textToWrite = strings.TrimSpace(textToWrite)

	lw.sourceText = textToWrite

	// prepare lines to write
	lw.linesToWrite = ConvertStringToLines(lw.sourceText, lw.fontFace, lw.maxLineWidth)
	if len(lw.linesToWrite) == 0 {
		logz.Panicln("LineWriter", "no lines to write. source text:", lw.sourceText)
	}
	lw.currentLineIndex = 0
	lw.currentLineNumber = 0
	lw.writtenLines = []string{""}
	lw.lineHeight = 0

	// determine line height
	for _, line := range lw.linesToWrite {
		_, lineHeight, _ := GetStringSize(line, lw.fontFace)
		if lineHeight > lw.lineHeight {
			lw.lineHeight = lineHeight
		}
	}

	if lw.lineHeight == 0 {
		logz.Panicln("LineWriter", "lineheight was 0. linesToWrite:", lw.linesToWrite)
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

	// we set this here since cursorY needs to be initialized to lineHeight, but this is the place that calculates that.
	lw.resetCursor()

	lw.WritingStatus = Writing
}

// Clear fully clears source text and all written text
func (lw *LineWriter) Clear() {
	if lw.WritingStatus == Writing {
		panic("tried to clear linewriter while text is being written!")
	}

	// unset all the values that are calculated when setting source text
	lw.linesToWrite = []string{}
	lw.currentLineIndex = 0
	lw.currentLineNumber = 0
	lw.writtenLines = []string{}
	lw.lineHeight = 0
	lw.currentPage = 0
	lw.pages = make([][]string, 0)
	lw.pageLineCount = 0

	// Note: we don't reset the cursor here, since that needs to know the line height of the text.
	lw.textImg.Clear()

	// set status flag to indicate waiting for new text
	lw.WritingStatus = AwaitText
}

func (lw *LineWriter) resetCursor() {
	if lw.lineHeight == 0 {
		panic("lineHeight was 0")
	}
	lw.cursorX = lw.cursorOffsetX
	lw.cursorY = lw.lineHeight
	lw.currentFgColor = lw.fgColor
	lw.currentShadow = lw.shadow
	lw.currentLinkPos = nil
}

// Draw returns the last written Y position (for reference by other drawing functions)
func (lw LineWriter) Draw(screen *ebiten.Image, startX, startY int) int {
	if lw.textImg == nil {
		panic("textImg was nil")
	}
	rendering.DrawImage(screen, lw.textImg, float64(startX), float64(startY), 0)

	return lw.cursorY
}

func (lw *LineWriter) drawRune(r, prev rune) {
	if lw.textImg == nil {
		panic("text img is nil")
	}
	if lw.sourceText == "" {
		panic("source text hasn't been set yet")
	}
	if lw.lineHeight == 0 {
		panic("lineHeight was 0")
	}
	if lw.WritingStatus != Writing {
		logz.Panicln("LineWriter", "tried to draw text, but linewriter status isn't set to Writing. status:", lw.WritingStatus)
	}

	_, dy, _ := GetStringSize(string(r), lw.fontFace)
	dx := AdvanceWidth(r, prev, lw.fontFace)

	// sanity check: make sure the cursor isn't drawing outside the bounds of textImage
	if lw.cursorY-dy < 0 {
		logz.Println("LineWriter", "r:", r)
		logz.Panicln("LineWriter", "it looks like the cursor is drawing text above the text box (clipping it at the top)", "cursorY:", lw.cursorY, "dy:", dy)
	}
	if lw.cursorY > lw.maxHeight {
		logz.Println("LineWriter", "r:", r)
		logz.Panicln("LineWriter", "it looks like the cursor is drawing text below the text box (clipping it at the bottom)", "cursorY:", lw.cursorY, "maxHeight:", lw.maxHeight)
	}
	if lw.cursorX < 0 {
		logz.Println("LineWriter", "r:", r)
		logz.Panicln("LineWriter", "it looks like the cursor is drawing text left of the text box (clipping it at the beginning)", "cursorX:", lw.cursorX)
	}
	if lw.cursorX+dx > lw.maxLineWidth+lw.cursorOffsetX {
		logz.Println("LineWriter", "r:", r)
		logz.Panicln("LineWriter", "it looks like the cursor is drawing text right of the text box (clipping it at the end)", "cursorX:", lw.cursorX, "dx:", dx, "maxWidth:", lw.maxLineWidth, fmt.Sprintf("s: \"%s\"", string(r)))
	}

	// apply kerning before draw
	kern := lw.fontFace.Kern(prev, r)
	lw.cursorX += kern.Round()

	if lw.currentShadow {
		DrawShadowText(lw.textImg, string(r), lw.fontFace, lw.cursorX, lw.cursorY, lw.currentFgColor, lw.bgColor, -2, -2)
	} else {
		DrawText(lw.textImg, string(r), lw.fontFace, lw.cursorX, lw.cursorY, lw.currentFgColor)
	}

	adv, ok := lw.fontFace.GlyphAdvance(r)
	if ok {
		// this should pretty much never fail, but I guess some glyphs don't have advance in theory
		lw.cursorX += adv.Round()
	}

	// assert: dx = kern + adv, just to make sure nothing funky is going on
	if dx != (kern + adv).Round() {
		logz.Panicln("drawRune", "AdvanceWidth didn't add up to kern + advance... something must've gone wrong. dx:", dx, "kern:", kern.Round(), "adv:", adv.Round())
	}
}

func (lw *LineWriter) newLine() {
	if lw.lineHeight == 0 {
		panic("lineHeight was 0")
	}
	lw.cursorX = lw.cursorOffsetX
	lw.cursorY += lw.lineHeight
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

			// ensure runes slice has been made for this line.
			// Note: the reason we can't use bytes is because some languages' characters are represented by multiple bytes.
			// (e.g. "A" is 1 byte, but "あ" is 3 bytes)
			// so, if we need to draw runes, we need to get the runes from a string. Getting a byte at a certain index will not work
			// unless we are only dealing with ASCII single byte characters.
			if len(lw.currentLineRunes) == 0 {
				for _, r := range currentPage[lw.currentLineNumber] {
					lw.currentLineRunes = append(lw.currentLineRunes, r)
				}
			}

			if lw.currentLineIndex < len(lw.currentLineRunes) {
				lw.textUpdateTimer++

				if lw.textUpdateTimer >= ticksPerTextWrite+lw.addedUpdateTimer {
					// continue to write the current line
					next := lw.currentLineRunes[lw.currentLineIndex]
					prev := rune(0)
					if lw.currentLineIndex > 0 {
						prev = lw.currentLineRunes[lw.currentLineIndex-1]
					}

					skippedRune := lw.handleDrawRune(next, prev)

					if !skippedRune {
						lw.textUpdateTimer = 0

						// see if we should slow down the writing due to an upcoming end of sentence
						if isSlowdownChar(next) {
							lw.addedUpdateTimer = 25
						} else {
							lw.addedUpdateTimer = 0

							// play sfx when we aren't drawing a slow-down character. it sounds weird if the sfx plays randomly during series of "..."
							if lw.textBlipSfx != "" {
								if lw.audioman == nil {
									panic("audioman was nil")
								}
								lw.ticksTilBlip--
								if lw.ticksTilBlip <= 0 {
									lw.audioman.PlaySFX(lw.textBlipSfx, 0.2)
									lw.ticksTilBlip = lw.blipTickInterval
								}
							}
						}
					}

					// update line index
					lw.writtenLines[lw.currentLineNumber] += string(next)
					lw.currentLineIndex++
				}
			} else {
				// go to the next line
				lw.currentLineNumber++
				lw.currentLineIndex = 0
				lw.newLine()
				lw.writtenLines = append(lw.writtenLines, "") // start next output line
				lw.currentLineRunes = []rune{}                // reset runes
			}
		} else {
			// everything has been written for the current page
			lw.currentLineRunes = []rune{}
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

func (lw *LineWriter) handleDrawRune(next, prev rune) (skipThisRune bool) {
	if lw.currentFgColor == nil {
		lw.currentFgColor = lw.fgColor
	}
	// check for symbols that may change fg color
	skipThisRune = false
	if lw.specialSymbols {
		switch next {
		case '[':
			// start a new link position
			if lw.currentLinkPos != nil {
				logz.Panic("hit a [, but there's already a current link pos. if a previous link position was finished, this pointer needs to be cleared. otherwise, there is a [ inside an already opened link.")
			}
			lw.currentLinkPos = &LinkPos{
				X:          float64(lw.cursorX),
				Y:          float64(lw.cursorY),
				LineHeight: lw.lineHeight,
			}
			lw.currentFgColor = lw.linkColor
			skipThisRune = true
		case ']':
			// finish existing link position
			if lw.currentLinkPos == nil {
				logz.Panic("hit a ], but no current link pos is defined. is there a stray ] sitting in dialog text? or did the open [ not start a new current link pos?")
			}
			lw.currentLinkPos.X2 = float64(lw.cursorX)
			lw.currentLinkPos.Y2 = float64(lw.cursorY)
			lw.linkPositions = append(lw.linkPositions, *lw.currentLinkPos)
			lw.currentLinkPos = nil
			lw.currentFgColor = lw.fgColor
			skipThisRune = true
		case '_':
			if ColorsEqual(lw.currentFgColor, lw.bgColor) {
				lw.currentFgColor = lw.fgColor
				lw.currentShadow = lw.shadow
			} else {
				lw.currentFgColor = lw.bgColor
				lw.currentShadow = false
			}
			skipThisRune = true
		}
	}

	if !skipThisRune {
		lw.drawRune(next, prev)
	}

	return skipThisRune
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

	lw.linkPositions = []LinkPos{}

	lw.resetCursor()
	lw.textImg.Clear()

	lw.currentPage++
	lw.WritingStatus = Writing
}

// FastForward finishes the current page
func (lw *LineWriter) FastForward() {
	currentPage := lw.pages[lw.currentPage]
	lw.writtenLines = []string{}
	lw.writtenLines = append(lw.writtenLines, currentPage...)
	lw.linkPositions = []LinkPos{}

	lw.resetCursor()
	lw.textImg.Clear()
	for _, line := range currentPage {
		// Note: we draw character by character since, soon, I plan to add logic that checks an individual character and changes draw behavior
		// e.g. drawing asides in a different color by using underscores
		// even though this looks pretty inefficient, realistically I don't think it'll have any impact
		prev := rune(0)

		for _, r := range line {
			lw.handleDrawRune(r, prev)
			prev = r
		}

		lw.newLine()
	}

	lw.currentLineNumber = len(currentPage)
	lw.currentLineIndex = 0
}

type LinkPos struct {
	X, Y       float64
	X2, Y2     float64
	LineHeight int
}

// GetLinkPositions returns the start and end positions of link text drawn in a lineWriter.
// Used for overlaying a button on top of link text so simulate a link for things like topics.
// Returned x and y are of the text writer's cursor, so y is from the bottom of the text, not the top like other things.
// Panics if called while lineWriter is busy writing.
func (lw LineWriter) GetLinkPositions() []LinkPos {
	if lw.WritingStatus == Writing {
		logz.Panic("tried to get link positions while linewriter was still writing")
	}
	return lw.linkPositions
}
