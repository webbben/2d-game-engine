# Guide to Tiled

Recording useful or important info about how to use the Tiled editor here, since it could get a bit complicated over time.

## Maps and Tilesets

Maps have tile layers, which is the basic type of layer for adding tiles.

a tile layer looks like this in the JSON:

```json
"layers":[
        {
         "data":[659, 659, 659, 288, 452, 452, 452, 452, 288, ...],
         ...
        },
]
```

Each number in "data" is a `gid` of a tile from a tileset.

### Tilesets

When a tileset is added to a map, you see it in the JSON like this:

```json
"tilesets":[
        {
         "firstgid":1,
         "source":"..\/tilesets\/sdv_outdoors_spring.tsj"
        }]
```

This defines the first `gid` of the first tile in this tileset.

> the range of gids is probably defined by the tile width and height of the tileset? just my guess.

Inside the actual tileset file, it is essentially just a link to a singular image file (the tileset image), and definition of the tile width/height. It also has all the configuration or metadata for things like animations.

> So, when we render a map that uses a certain tileset, we will need to:
>
> 1. load the image file of the tileset
> 2. split it up into all the individual tile images by tile width/height

# Our Implementation of Tiled Maps

Below I'll record decisions made about how we will use certain features of Tiled maps with this game.

## Collisions and "Cost" for path finding

Tiles will have a property called "cost" which represents how "difficult" it is to move through this tile.
This is used for things like a-star path finding algorithms, but we will also use it to effectively mark the collisions of a map too.

Let's say that if a tile has a `cost >= 10`, then it is officially a "collision" and cannot be passed through.
For the path finding algorithm, if a tile has a cost of 10 or higher, we will just bump that tile's cost to some arbitrarily high number like 9999, to ensure the path finding algorithm never tries to go through it.

## Using Tilesets for Entity Animation Definitions

We use Tiled tilesets to define all the frames for animations for entities.
We use the `frames` property to do this.

the `frames` property expects the following format:
`animationName1:0,animationName2:x`
where `animationName` is the name of a specific animation, and the number is the frame number in that animation

This way, you can reuse the same tile for more than one animation definition if you want.
