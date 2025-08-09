package room

type Map struct {
	BackgroundColor  string  `json:"backgroundcolor"`
	Class            string  `json:"class"`
	CompressionLevel int     `json:"compressionlevel"`
	Height           int     `json:"height"` // number of tile rows
	Width            int     `json:"width"`  // number of tile columns
	HexSideLength    int     `json:"hexsidelength"`
	Infinite         bool    `json:"infinite"`
	Layers           []Layer `json:"layers"`
	Orientation      string  `json:"orientation"`
	Properties       []any   `json:"properties"`
	RenderOrder      string  `json:"renderorder"`
	StaggerAxis      string  `json:"staggeraxis"`
	StaggerIndex     string  `json:"staggerindex"`
	TiledVersion     string  `json:"tiledversion"`
	TileHeight       int     `json:"tileheight"` // map grid height
	TileWidth        int     `json:"tilewidth"`  // map grid width
	Tilesets         []any   `json:"tilesets"`
}

type Layer struct {
	Chunks      []any   `json:"chunks"`
	Compression string  `json:"compression"` // zlib, gzip, zstd or empty (default)
	Data        []uint  `json:"data"`
	DrawOrder   string  `json:"draworder"`   // "topdown" or "index" (default). objectgroup only.
	Encoding    string  `json:"encoding"`    // "csv" (default) or "base64". tilelayer only.
	Height      int     `json:"height"`      // row count. same as map height for fixed-size maps. tilelayer only.
	ID          int     `json:"id"`          // incremental ID - unique across all layers
	Image       string  `json:"image"`       // image used by this layer. imagelayer only.
	ImageHeight int     `json:"imageheight"` // height of the image used by this layer. imagelayer only.
	ImageWidth  int     `json:"imagewidth"`  // width of the image used by this layer. imagelayer only.
	Layers      []Layer `json:"layers"`      // array of layers. group only.
	Locked      bool    `json:"locked"`      // whether layer is locked in the editor
	Name        string  `json:"name"`        // name assigned to this layer
	Objects     []any   `json:"objects"`
}
