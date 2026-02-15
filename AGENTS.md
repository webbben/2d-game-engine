# AGENTS.md

## Project Overview

This is a 2D game engine in Go using the Ebiten framework for RPG-style games. Provides entity management, rendering, UI, dialog, trading, and more.

## Build, Lint & Test Commands

### Running the Game
```bash
./run_game.sh
go run .
```

### Running Tests
```bash
go test ./...                    # all tests
go test -v ./...                 # verbose
go test -run TestFindPath ./internal/path_finding/...  # single test
go test -run "Test.*" ./...      # pattern match
go test -bench=. ./internal/path_finding/...           # benchmarks
go test -cover ./...             # coverage
```

### Linting & Formatting
```bash
go fmt ./...
go vet ./...
golangci-lint run ./...          # if available
```

### Building
```bash
go build -o bin/character-builder ./cli/main.go
go build -o game-binary .
GOOS=windows GOARCH=amd64 go build -o game.exe .  # cross-compile
```

## Code Style Guidelines

### Formatting
- Run `go fmt ./...` before committing
- Use tabs (Go standard)
- Keep lines under 100 characters
- Blank lines between functions and import groups

### Imports (3 groups, separated by blank lines)
1. Standard library
2. Third-party (github.com, etc.)
3. Internal packages
- Use aliases: `m "github.com/..."`

### Naming
- **Types/Functions**: PascalCase (`Player`, `NewGame()`)
- **Variables/Fields**: camelCase (`playerName`, `currentHealth`)
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE
- **Packages**: lowercase (`dialog`, `trade`)
- **Files**: lowercase with underscores (`path_finding.go`)

### Error Handling
- Return `error` for recoverable failures (file I/O, user input)
- Use `panic` for programming errors or unrecoverable states
- Validate inputs in public functions
- Wrap errors: `fmt.Errorf("failed to load: %w", err)`

### Types & Interfaces
- Define custom types near first usage
- Use interfaces for abstraction; implement implicitly
- Prefer concrete types internally; interfaces at boundaries
- Use pointers (`*Type`) for large structs or mutation

### Documentation
- Document exported types/functions with doc comments
- Start comments with identifier name: `// Player represents...`
- Keep concise; document non-obvious behavior

### Testing
- Tests in `*_test.go` alongside source
- Use table-driven tests
- Name: `Test<FunctionName>`, use `t.Run`
- Include `Benchmark*` for performance-critical code

## Architecture

### Key Packages
- `game/` - Core game loop, camera, time system, events
- `entity/` - Base entity, player, NPCs, body system
- `item/` - Item definitions, inventory items
- `internal/ui/` - Buttons, text fields, dropdowns, layout
- `dialogv2/` - Topic-based conversations
- `quest/` - Event-driven quest system (stage-based progression)
- `trade/` - Player-merchant trading
- `definitions/` - JSON-based game data
- `data/defs/` - Type definitions for dialog, quests, items, etc.
- `data/state/` - Runtime state structures

### Patterns
- Constructor: `New[Type]()` with validation
- Interface-based: `ItemDef`, `WorldContext`, `Renderable`
- Fail-fast: panic on invalid internal state

## Common Tasks

### Adding Items
1. Define in JSON
2. Implement `ItemDef` if custom behavior needed
3. Add to DefinitionManager

### Creating NPCs
1. Define entity data in definitions
2. Configure AI tasks/behaviors
3. Set up dialog topics
4. Place in Tiled map

## Performance
- Object pooling for frequent entity creation/destruction
- Implement culling for off-screen objects
- Cache rendered images
- Spatial partitioning for collision detection

## Documentation

### _docs Directory
New features and systems should be documented in the `_docs/` directory. See existing docs:
- `_docs/dialog_system.md` - Topic-based conversation system
- `_docs/quest_system.md` - Event-driven quest progression

### Documentation Guidelines
- Explain high-level architecture and design decisions
- Include code examples for core types
- Document edge cases and important behaviors
- Keep docs in sync with implementation
