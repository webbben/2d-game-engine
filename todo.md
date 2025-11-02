# TODO

Tracking ongoing tasks and their progress

## Dialog

- [ ] Create new dialog framework
  - [x] display text in dialog boxes
  - [ ] add support for dialog options, clickable by a mouse
  - [ ] create a fully fleshed dialog experience, with options, actions, and exiting the dialog

## Entity

- [x] design system of combining reusable smaller images into full frames
  - this will enable us to easily create varied entities without creating entire individual frame images
  - [ ] plug new system of reusable entity components into the actual entity drawing logic

## Rendering

# Useful resources

## Tilesets

- https://www.spriters-resource.com/pc_computer/stardewvalley/

# Done

## Entity

- [x] rewrite logic for moving an entity between tiles
- [x] plan how to design an entity and define its movement frames efficiently
- [x] test entity movement

## Incorporate Tiled format

- [x] get format of Tiled maps into a struct
- [x] refactor code for Rooms to render the Tiled format

## Rendering

- [x] add support for animated tiles (like grass tiles that have passive animations)

## Physics

- [x] Finish collisions
  - [x] player <-> entity
  - [x] objects, structures
