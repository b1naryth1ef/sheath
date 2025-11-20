package ecs_test

import "github.com/b1naryth1ef/sheath/ecs"

// Common test component types
type Position struct {
	X, Y float32
}

type Velocity struct {
	DX, DY float32
}

type Name struct {
	Value string
}

type Health struct {
	Current int
	Max     int
}

type PlayerController struct{}

type AI struct {
	State int
}

// Custom primitive types for testing non-pointer components
type Score int32
type Tag string
type Temperature float64

type TestA string
type TestB string

type AIPointer struct {
	Target *Position
}
type Inventory struct {
	Items []string
}
type Stats struct {
	Attributes map[string]int
}
type Target struct {
	Enemy *Name
}
type Link struct {
	Next *Position
}
type Inner struct {
	Value int
}
type Outer struct {
	Data *Inner
	List []*Inner
}
type RefComponent struct {
	Ref *Position
}

func registerTestComponents() {
	ecs.RegisterComponent[Position]()
	ecs.RegisterComponent[Velocity]()
	ecs.RegisterComponent[Name]()
	ecs.RegisterComponent[Health]()
	ecs.RegisterComponent[PlayerController]()
	ecs.RegisterComponent[AI]()
	ecs.RegisterComponent[Score]()
	ecs.RegisterComponent[Tag]()
	ecs.RegisterComponent[Temperature]()
	ecs.RegisterComponent[TestA]()
	ecs.RegisterComponent[TestB]()
	ecs.RegisterComponent[int32]()
	ecs.RegisterComponent[float64]()
	ecs.RegisterComponent[string]()
	ecs.RegisterComponent[AIPointer]()
	ecs.RegisterComponent[Inventory]()
	ecs.RegisterComponent[Stats]()
	ecs.RegisterComponent[Target]()
	ecs.RegisterComponent[Link]()
	ecs.RegisterComponent[Inner]()
	ecs.RegisterComponent[Outer]()
	ecs.RegisterComponent[RefComponent]()
}
