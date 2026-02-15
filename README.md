# Ben's 2d Game Engine

This is my little pet project to build a game engine/future game. It's a 2d game engine written with Golang and mainly using the ebiten engine.

Essentially, this is a wrapper of ebiten, and here is the breakdown of "this engine" vs "ebitengine":

- Ebiten(gine):
  - the really low level aspects of a game engine. You could call it the "core infrastructure" that we are building on top of.
  - handles drawing things to a screen, and sets up the Update and Draw loops and manages/runs those.
  - all the stuff that I'm probably not really smart enough to do (or _certainly_ don't have time for. it's incredible what ebiten does.)

- This game engine:
  - wrapper on top of ebiten, and essentially defines all the actual game concepts that we want to happen.
  - handles all runtime logic for all of these game concepts. ex: handles running dialog sessions, managing the data related to them, etc.
  - provides a "framework" that you can simply use to "plug and play" (define things, start the game, and all actual logic is hidden in this game engine).

So, I think it skirts the boundary of "game engine" because, well, ebiten is of course the "core" game engine. But instead of building my games directly tied to ebiten,
I wanted an additional layer of abstraction that is highly reusable, which will let me make games a lot faster in the long run.

## Usage 

This package can be imported into another go project, and the general idea is this:

1) The "game" project only defines data and definitions; it probably will not be writing much code that is injected into the Update or Draw loops.

- This package provides a bunch of definitions for things like dialog, quests, etc in the `data/defs` package.
- Only certain things, like UI screens, settings pages, etc will be directly "programmed" in the game project. All the lower level runtime logic for 
things like combat, movement, loading maps, loading and processing dialogs, quests, etc is all handled by the engine.

2) We may expose some hooks into the update or draw loops, but those should be used as sparingly as possible and only when absolutely necessary.

- for anything you'd want to create or draw on the screen, there's probably already a system for it in this engine.
- if there isn't, we should consider adding it to the engine.
- I'd really only say to use these "update loop hooks" for something that is very unique and a super-duper edge case for your game.

So, a completed game will probably look something like this:

- A bunch of files of go code that has a ton of constants or variables that define the definitions for characters, dialog profiles, quests, etc.
  - you can also put it all in JSON, but I personally like to put it directly into go code so the compilers and IDE can immediately detect syntax or other issues.
- All of the assets you use in the game, such as images, audio, fonts, etc.
- Code for creating the title screen, all the settings screens, and all in-game UI screens (player inventory page, player stats pages, etc).
  - this engine provides all the concepts needed for building these pages, such as UI components and access to all the relevant data via the root "Game" struct for ebiten, etc.
- Some amount of code to handle wrapping it all together, loading the definitions into the engine, and starting the game.

## Status Overview 

I'll track the overall status of how I (think) each of these systems are functioning, and/or my plans to develop them in the future:

### Finished or "basically finished"

- Items
  - I'm not aware of any missing things here, but that can change of course.
- Dialog 
  - recently rewrote it to be much more robust and support a much more flexible system. seems "done" to me but of course I haven't tested much yet.
- UI components (for screens, etc)
  - pretty much done I'd say. I add new components from time to time, or enhance existing ones, but by now we have a lot of stuff that can be used to make complex screens.

### Getting close to finished 

- Maps 
  - loading tiled maps, and rendering entities in them.
  - TODO: probably some things about managing moving between maps, the "game world" as a concept, etc.

- Entity core logic:
  - movement in a map, collisions, etc.
  - body rendering and animation system 
  - TODO: probably things like handling combat

- Quests 
  - recently wrote this system, but haven't tested yet. But, it feels like it's close to being fully functional.

- Trading system 
  - seems to work pretty well, but I haven't touched this in a while and a big refactor occurred, so I should test this more and make sure it's not missing anything.


### WIP 

- Combat system 
  - have some basic functions created for handling attacks and damaging enemies, but it's still in its infancy and could change.

- Skills, Attributes, and Leveling system 
  - skills and attributes have been created as a concept, and I'm made the formula for calculating characters' levels based on their skills.
  - haven't yet made a way to have skills and attributes influence things like combat, or other mechanics that they should be able to influence.


### Future (Not Started)

## Dev Log 

I have a dev log in the _docs directory, which I use to track how things are going, talk about issues or new things I'm working on, etc.
Since this engine is still in kind of an early phase of development, it mainly has discussions on technical issues, brainstorming new game concepts, and giving updates 
on things like the "Character Builder" utility. But it will eventually also be a place to discuss the game I'm working on.

[Dev Log](/_docs/devlog/devlog.md)

