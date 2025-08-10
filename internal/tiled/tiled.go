package tiled

import "encoding/json"

// Map represents the root map structure from Tiled
type Map struct {
	Type             string     `json:"type"`
	Version          string     `json:"version"`
	TiledVersion     string     `json:"tiledversion"`
	Orientation      string     `json:"orientation"`           // orthogonal, isometric, staggered, hexagonal
	RenderOrder      string     `json:"renderorder,omitempty"` // right-down, right-up, left-down, left-up
	Width            int        `json:"width"`
	Height           int        `json:"height"`
	TileWidth        int        `json:"tilewidth"`
	TileHeight       int        `json:"tileheight"`
	BackgroundColor  string     `json:"backgroundcolor,omitempty"` // Hex color
	Layers           []Layer    `json:"layers"`
	Tilesets         []Tileset  `json:"tilesets"`
	Properties       []Property `json:"properties,omitempty"`
	NextLayerID      int        `json:"nextlayerid,omitempty"`
	NextObjectID     int        `json:"nextobjectid,omitempty"`
	ParallaxOriginX  float64    `json:"parallaxoriginx,omitempty"`
	ParallaxOriginY  float64    `json:"parallaxoriginy,omitempty"`
	HexSideLength    int        `json:"hexsidelength,omitempty"`
	StaggerAxis      string     `json:"staggeraxis,omitempty"`  // x, y
	StaggerIndex     string     `json:"staggerindex,omitempty"` // odd, even
	Infinite         bool       `json:"infinite,omitempty"`
	CompressionLevel int        `json:"compressionlevel,omitempty"`
}

// Layer represents a layer in the map
type Layer struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Type       string     `json:"type"` // tilelayer, objectgroup, imagelayer, group
	Visible    bool       `json:"visible"`
	Opacity    float64    `json:"opacity"`
	OffsetX    float64    `json:"offsetx,omitempty"`
	OffsetY    float64    `json:"offsety,omitempty"`
	ParallaxX  float64    `json:"parallaxx,omitempty"`
	ParallaxY  float64    `json:"parallaxy,omitempty"`
	Properties []Property `json:"properties,omitempty"`

	// Tile layer specific
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Data        []int  `json:"data,omitempty"`        // Tile data as array of tile IDs
	Encoding    string `json:"encoding,omitempty"`    // base64
	Compression string `json:"compression,omitempty"` // zlib, gzip, zstd

	// Object layer specific
	Objects   []Object `json:"objects,omitempty"`
	DrawOrder string   `json:"draworder,omitempty"` // topdown, index

	// Image layer specific
	Image string `json:"image,omitempty"`

	// Group layer specific
	Layers []Layer `json:"layers,omitempty"`

	// Common optional fields
	TintColor string `json:"tintcolor,omitempty"`
	Class     string `json:"class,omitempty"`
}

// Object represents an object in an object layer
type Object struct {
	ID         int        `json:"id"`
	Name       string     `json:"name,omitempty"`
	Type       string     `json:"type,omitempty"`
	Class      string     `json:"class,omitempty"`
	X          float64    `json:"x"`
	Y          float64    `json:"y"`
	Width      float64    `json:"width,omitempty"`
	Height     float64    `json:"height,omitempty"`
	Rotation   float64    `json:"rotation,omitempty"`
	Visible    bool       `json:"visible"`
	Properties []Property `json:"properties,omitempty"`

	// Shape-specific fields
	Ellipse  bool    `json:"ellipse,omitempty"`  // true if ellipse
	Point    bool    `json:"point,omitempty"`    // true if point
	Polygon  []Point `json:"polygon,omitempty"`  // Array of points for polygon
	Polyline []Point `json:"polyline,omitempty"` // Array of points for polyline

	// Tile object specific
	GID int `json:"gid,omitempty"` // Global tile ID

	// Text object specific
	Text *Text `json:"text,omitempty"`

	// Template reference
	Template string `json:"template,omitempty"`
}

// Point represents a coordinate point
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Text represents text object properties
type Text struct {
	Text       string `json:"text"`
	FontFamily string `json:"fontfamily,omitempty"`
	PixelSize  int    `json:"pixelsize,omitempty"`
	Wrap       bool   `json:"wrap,omitempty"`
	Color      string `json:"color,omitempty"`
	Bold       bool   `json:"bold,omitempty"`
	Italic     bool   `json:"italic,omitempty"`
	Underline  bool   `json:"underline,omitempty"`
	Strikeout  bool   `json:"strikeout,omitempty"`
	Kerning    bool   `json:"kerning,omitempty"`
	HAlign     string `json:"halign,omitempty"` // left, center, right, justify
	VAlign     string `json:"valign,omitempty"` // top, center, bottom
}

// Tileset represents a tileset reference
type Tileset struct {
	FirstGID int    `json:"firstgid"`
	Source   string `json:"source,omitempty"` // External tileset file

	// Embedded tileset properties
	Name        string      `json:"name,omitempty"`
	TileWidth   int         `json:"tilewidth,omitempty"`
	TileHeight  int         `json:"tileheight,omitempty"`
	TileCount   int         `json:"tilecount,omitempty"`
	Columns     int         `json:"columns,omitempty"`
	Image       string      `json:"image,omitempty"`
	ImageWidth  int         `json:"imagewidth,omitempty"`
	ImageHeight int         `json:"imageheight,omitempty"`
	Margin      int         `json:"margin,omitempty"`
	Spacing     int         `json:"spacing,omitempty"`
	Properties  []Property  `json:"properties,omitempty"`
	Tiles       []Tile      `json:"tiles,omitempty"`
	TileOffset  *TileOffset `json:"tileoffset,omitempty"`
	Grid        *Grid       `json:"grid,omitempty"`

	// Background color
	BackgroundColor string `json:"backgroundcolor,omitempty"`

	// Transformation flags
	TransformationType string `json:"type,omitempty"`
	Class              string `json:"class,omitempty"`
}

// Tile represents individual tile properties within a tileset
type Tile struct {
	ID          int        `json:"id"`
	Type        string     `json:"type,omitempty"`
	Class       string     `json:"class,omitempty"`
	Properties  []Property `json:"properties,omitempty"`
	Image       string     `json:"image,omitempty"`
	ImageWidth  int        `json:"imagewidth,omitempty"`
	ImageHeight int        `json:"imageheight,omitempty"`
	Animation   []Frame    `json:"animation,omitempty"`
	ObjectGroup *Layer     `json:"objectgroup,omitempty"` // Collision shapes
	Terrain     []int      `json:"terrain,omitempty"`     // Deprecated
	Probability float64    `json:"probability,omitempty"`
}

// Frame represents an animation frame
type Frame struct {
	TileID   int `json:"tileid"`
	Duration int `json:"duration"` // milliseconds
}

// TileOffset represents tile rendering offset
type TileOffset struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Grid represents grid settings for isometric tilesets
type Grid struct {
	Orientation string `json:"orientation"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

// Property represents a custom property
type Property struct {
	Name         string      `json:"name"`
	Type         string      `json:"type,omitempty"` // string, int, float, bool, color, file, object, class
	Value        interface{} `json:"value"`
	PropertyType string      `json:"propertytype,omitempty"` // For object/class types
}

// UnmarshalJSON handles the flexible property value types
func (p *Property) UnmarshalJSON(data []byte) error {
	type Alias Property
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Set default type if not specified
	if p.Type == "" {
		p.Type = "string"
	}

	return nil
}

// GetStringValue returns the property value as a string
func (p *Property) GetStringValue() string {
	if str, ok := p.Value.(string); ok {
		return str
	}
	return ""
}

// GetIntValue returns the property value as an int
func (p *Property) GetIntValue() int {
	switch v := p.Value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return 0
}

// GetFloatValue returns the property value as a float64
func (p *Property) GetFloatValue() float64 {
	switch v := p.Value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	return 0.0
}

// GetBoolValue returns the property value as a bool
func (p *Property) GetBoolValue() bool {
	if b, ok := p.Value.(bool); ok {
		return b
	}
	return false
}
