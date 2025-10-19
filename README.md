# Ben's 2d Game Engine

This is my little pet project to build a game engine/future game. It's a 2d game engine written with Golang and mainly using the ebiten engine.

Still in very early stages of development, but as I develop it I'm gradually getting a better sense of what the finished product will look like. Here is the current concept and how I expect this to be used:

## Usage as a Game Engine

This module will contain all (or 90%) of the core game logic. So, for example, it handles basic stuff like entity movement, collisions, rendering maps, etc, but it also will have logic coded into it for the more specific mechanics of the game, like how to handle trading sessions, item definitions, combat, etc. Ideally, it is a flexible engine that takes care of almost all the programming aspects, and when you actually make the "game" part, you are mostly plugging in the game data like tilesets, maps, entity definitions, item definitions, etc. However, I think most of the programming that will be done outside of the game engine will be for defining the various screens, like the trading screen, inventory screens, etc. These will mainly be about drawing the screens using existing UI components, and calling the right functions to update the player's items, add money or take money from the player, etc.

So, theoretically, a "game" that uses this "game engine" would be made like this:

1. create a new go project
2. install this game engine as a dependency to your go project
3. define all the game data (entities, items, maps, etc) in your go project, using the correct packages from the game engine module. (e.g. import the `entity` package and define a slice of entities that will exist in the game world).
4. also, write some custom logic that is injected into the game's update loop via some hooks the game engine provides. this can be used for controlling things like showing certain UI menus, the player's inventory, etc.
5. Run the game!

### Packaging the game for distribution

In the end, we will need a binary that has access to the following:

1. the assets files, maps, tilesets, etc. (saved in a game files directory)
2. the game data such as items, entities, etc (directly coded into the go project)

If using Steam, the files from `(1)` should be placed in a game-specific directory by steam.
So, you will just need to ensure that the game's assets directory is correctly found in the game project, and the `config` variable is set appropriately. The game data such as items, etc will be directly embedded as part of the built binary.
