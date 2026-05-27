package config

import (
	"image/color"

	"github.com/webbben/2d-game-engine/data/defs"
	"golang.org/x/image/font"
)

/*
*	I've started moving "Params" structs into the config package since in some cases I want to reference them here in config.
*
*	Here are the reasons, generally speaking, why I think it works to put them here:
*		- "Params" are pretty much config, if you think about it (usually about how to configure certain UI components)
*		- most Params don't need to reference anything besides def IDs, primitive types (strings, ints, etc), or other Param types (i.e. when a UI component has another one nested in it)
*
 */

type LineWriterParams struct {
	LineWidthPx, MaxHeightPx int
	FontFace                 font.Face
	SupportSpecialSymbols    bool // if true, special symbols (like underscore, square brackets) will produce text formatting effects

	UseShadow            bool // if true, text will draw a shadow behind it using the BgColor
	WriteImmediately     bool // if true, text will write immediately on first update
	TextBlipSfx          defs.SoundID
	TextBlipTickInterval int

	FgColor   color.Color // color used for text foreground. defaults to black.
	BgColor   color.Color // color used for shadow behind text, or for de-emphasized/aside text (text inside underscores). defaults to light gray.
	LinkColor color.Color // color used for link text (text inside square brackets). defaults to blue.
}

type BookSessionParams struct {
	BoxTileset string
	BoxOrigin  int

	TitleFont font.Face

	LineWriterParams LineWriterParams
}
