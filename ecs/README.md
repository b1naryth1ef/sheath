# ECS

## Example

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
