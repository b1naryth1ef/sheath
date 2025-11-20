# ECS

## Examples

### Basic Usage

```go
package main

import (
	"github.com/b1naryth1ef/sheath/ecs"
)

type Position struct {
	X float64
	Y float64
}

type Health struct {
	Current uint8
	Max uint8
}

type PlayerName string

func main() {
	// Create a component registry and register all component types
	registry := ecs.NewComponentRegistry()
	ecs.RegisterComponent[Position](registry)
	ecs.RegisterComponent[Health](registry)
	ecs.RegisterComponent[PlayerName](registry)

	// Storage handles our entity/component storage
	storage := ecs.NewStorage(registry)

	// The easiest way to add a new entity is with the Spawn function
	player := storage.Spawn(
		Position{X: 3.5, Y: 5.3},
		Health{Current: 10, Max: 10},
		PlayerName("joe"),
	)

	// We can read components easily using the ReadComponent helper
	position := ecs.ReadComponent[Position](storage, player)

	// Since position is a pointer we can modify its contents here
	position.X += 5.5
	position.Y += 5.5
}
```

### Views

While the methods on `Storage` are nice to use for single entities, they can be inefficient for batch operations involving groups of entities that share similar components. We also often want to query all entities that contain a set of components. Both of these operations can be handled by a `View`, which can be saved and re-used for performance gains.

```go
type Targetable struct {
	*Position
	*Health
}

// Views use a struct which contains pointers to components
targetable := ecs.NewView[Targetable](storage)

// Now we can very efficiently spawn entities	
enemyId := targetable.Spawn(Targetable{
	Position: &Position{X: 100, Y: 100},
	Health: &Health{Current: 5, Max: 15},
})

// And views allow us clean and easy access to multiple components at a time
enemy := targetable.Get(enemyId)

// And now we have access to its components
enemy.Position.X = 200
enemy.Position.Y = 200

// Views are also used for querying multiple entities
for entityId, target := range targetable.Iter() {
	target.Health.Current -= 1
}

// And they support optional components for querying too
type CustomView struct {
	*Position
	*PlayerName `ecs:"optional"`
}

customView := ecs.NewView[CustomView](storage)

// We get all entities that have Position, PlayerName is optionally filled if available
for entityId, custom := range customView.Iter() {
	if custom.PlayerName != nil {
		// We have PlayerName
	}
}
```

### Archetypes

Archetypes are an abstraction which allow us to group entity storage based on their components. Generally you shouldn't have to worry about archetypes as the `Storage` layer abstracts all of this. However sometimes its useful to directly interact with archetypes, usually for efficiency purposes. To do this we can use the `Storage.GetArchetype()` and `Storage.GetArchetypeByTypes()` methods:

```go
type Player struct {}

ecs.RegisterComponent[Player](registry)
storage.Spawn(Player{})
archetype := storage.GetArchetype(Player{})
// or using reflect types:
// archetype := storage.GetArchetypeByTypes([]reflect.Type{reflect.TypeOf(Player{})})
```

### EntityId and Keeping References

It's tempting to use an `EntityId` to store references between entities. However this can be unsafe depending on how you interact with the `Storage` API. If you use the `Storage.AddComponent(...)`, `Storage.RemoveComponent(...)`, or `Archetype.Compact()` APIs, then `EntityId`'s are **not** guaranteed to be stable, and thus cannot be stored and used in-between these operations.

This is because an EntityId encodes both the archetype for the entity, and its position within the archetypes storage. This makes querying and using EntityId's extremely fast, but means they do not remain stable across operations that shuffle the underlying archetype or storage.

Instead we need to use an `EntityRef` which provides a safe and stable reference to an entity. Even as the underlying components of the entity change, or as it moves based on compaction, this reference will remain stable. Entity references are also weak, meaning we can detect when a reference has gone stale due to the underlying entity data being deleted.

```go
type AITarget struct {
	Ref ecs.EntityRef
}
type Player struct {}
type Misc struct {}

ecs.RegisterComponent[AITarget](registry)
ecs.RegisterComponent[Player](registry)
ecs.RegisterComponent[Misc](registry)

playerId := storage.Spawn(Player{})
playerRef := storage.CreateEntityRef(playerId)
aiId := storage.Spawn(AITarget{Ref: playerRef})

storage.AddComponent(playerId, Misc{})

// Safely get the entity despite it having moved around, and its EntityId changing
playerId, ok := storage.ResolveEntityRef(playerRef)
if ok {
	ai := ecs.ReadComponent[AITarget](storage, aiId)
	targetId, ok := storage.ResolveEntityRef(ai.Ref)
	if ok {
		// targetId is valid
	}
}
```

### Adding and Removing Components

Components can be added or removed from entities after they are spawned. Note that these operations may change the entity's archetype and invalidate the EntityId.

```go
type Velocity struct {
	DX, DY float64
}

ecs.RegisterComponent[Velocity](registry)

// Add a component to an existing entity
id := storage.Spawn(Position{X: 0, Y: 0})
newId := storage.AddComponent(id, Velocity{DX: 1.0, DY: 1.0})

// Remove a component from an entity
newId = storage.RemoveComponent(newId, reflect.TypeOf(Velocity{}))
```

### Compaction

The underlying storage for components groups them into blocks of 256. This provides good cache efficiency and also lets us easily track slot status via bitmasks. While it's very unlikely you run into fragmentation issues, it is possible for archetypes that experience frequent spawning/despawning with a small percentage of entities remaining long-term. In this case it'd be very likely we experience heavy fragmentation. We can optionally call the `Archetype.Compact()` method to forcefully compact the storage, removing empty slots and reducing the number of blocks allocated. Compaction will move entity storage and thus may invalidate existing EntityId's.

```go
type Player struct {}

storage.Spawn(Player{})
archetype := storage.GetArchetype(Player{})
archetype.Compact()
```
