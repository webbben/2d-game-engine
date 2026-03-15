# NPCs and Tasks

In this game engine, "NPCs" basically allow entities to have some logic and customized behavior.
Essentially a wrapper around an entity to provide the framework to control it and give it a "brain", more or less.

The main mechanism to control an NPC (and therefore its entity) is with Tasks. You can also just make your own custom Update logic and inject it into the NPC, too, but Tasks are useful as they have predefined tasks for common things like telling an NPC to go to a certain position, or telling it follow another entity, attack it, etc.

The Task framework also makes an NPC more reusable and customizeable; instead of hard-coding the logic for the NPC directly in an Update loop, you can define multiple different tasks which can be easily set, cancelled, reset, shared between NPCs, etc.

## Tasks

Tasks basically let you plug into a few different "lifecycle hooks" that define the task itself:

### Start

The `StartFn` is called when a task is starting. It could be used to initialize the NPC's state, set certain goals, etc.

### Update

the `OnUpdateFn` is called for a task on each Update loop (unless the task has already finished).
It is where you put most of the "meat" of the task logic. Mostly things that help the NPC proceed towards its goal.
Note that you don't need to put checks for success or failure, since there are separate hooks for these.

### Completion Check

The `IsCompleteFn` is called on each update to check if the task has met its success conditions.
So, this is where you define what the task's success is. For this reason, it's unnecessary to write these conditions into `OnUpdateFn`.

### Failure Check

Similar to the completion check, `IsFailureFn` is called on each update to check if the task has failed.
There isn't a functional difference between this and `IsCompleteFn` - just a semantic difference.
So, here you can define conditions for when the Task should fail early, before completion has been reached.

### End

This is a lifecycle hook for defining some "wrapping up" logic to the task. Mainly for setting some NPC/entity state, or something like that.
No task updates should occur here, because the End lifecycle hook also deactivates the task and disconnects it from its owner - so no more Update loops will occur after this lifecycle hook runs.

> To manually trigger a task to end, you can use the `npc.EndCurrentTask` function. This basically just tells the task to run its End lifecycle hook.
