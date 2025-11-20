# ECS

## Examples

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

type Targetable struct {
	*Position
	*Health
}

func main() {
	// Storage handles our entity/component storage
	storage := ecs.NewStorage()

	// The easiest way to add a new entity is with the Spawn function
	player := storage.Spawn(
		&Position{X: 3.5, Y: 5.3},
		&Health{Current: 10, Max: 10},
		PlayerName("joe"),
	)

	// We can read components easily using the ReadComponent helper
	position := ecs.ReadComponent[Position](storage, player)

	// Since position is a pointer we can modify its contents here
	position.X += 5.5
	position.Y += 5.5
	
	// These methods are relatively inefficient for general use however, so we instead want to use View's
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
	customView := ecs.NewView[struct {
		*Position
		*PlayerName `ecs:"optional"`
	}]

	// We get all entities that have Position, PlayerName is optionally filled if available
	for entityId, custom := range customView.Iter() {
		if custom.PlayerName != nil {
			// We have PlayerName
		}
	}
}
```

### Archetypes

Archetypes are an abstraction which allow us to group entity storage based on their components. Generally you shouldn't have to worry about archetypes as the `Storage` layer abstracts all of this. However sometimes its useful to directly interact with archetypes, usually for efficiency purposes. To do this we can use the `Storage.GetArchetype()` and `Storage.GetArchetypeByTypes()` methods:

```go
type Player struct {}

storage.Spawn(&Player{})
archetype := storage.GetArchetypeByTypes(reflect.TypeOf[Player]())
// archetype := storage.GetArchetype(&Player{})
```

### Entity References

The easiest way to reference other entities is by storing a `EntityId` in a component. This is always safe unless your code calls the `Archetype.Compact()` function. In this case you **cannot** safely store `EntityId`'s across calls to `Compact()`, instead you should use the `Storage.GetEntityRef(EntityId)` which returns a stable EntityRef that can be de-referenced via `Storage.ResolveEntityRef(EntityRef)`.

```go
type AITarget EntityId
type Player struct {}

playerId := storage.Spawn(&Player{})
storage.Spawn(AITarget(playerId))

// To safely use compact we must use EntityRefs
type AITargetSafe EntityRef
playerRef := storage.GetEntityRef(playerId)
aiId := storage.Spawn(AITargetSafe(playerRef))
```

### Compaction

The underlying storage for components groups them into blocks of 64. This provides good cache efficiency and also lets us easily track slot status via a `uint64`. While its very unlikely you run into fragementation issues, it is possible for archetypes that experience frequent spawning/despawning with a small percentage of entities remaining long-term. In this case it'd be very likely we experience heavy fragementation. We can optionally call the `Archetype.Compact()` method to forcefully compact the storage, removing empty slots and reducing the number of groups allocated. The downside to this is that `EntityId`'s will no-longer be stable after compaction. To safely reference entities in archetypes that may be compacted, please use `EntityRef`.

```go
type Player struct {}

storage.Spawn(&Player{})
archetype := storage.GetArchetypeByTypes(reflect.TypeOf[Player]())
archetype.Compact()
````
