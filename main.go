package main

import (
	"github.com/alexsem80/go-mapper/mapper"
)

func main() {
	src := Source{
		ID:     1,
		Name:   "Name",
		Weight: 32.32,
		Marks:  []int32{2, 3, 4},
		Address: []NestedSource{
			{
				State: "RYA",
				Index: 3900,
			},
			{
				State: "MOW",
				Index: 4700,
			},
		},
	}

	dest := &Destination{}

	mapper := mapper.NewMapper()
	mapper.CreateMap((*Source)(nil), (*Destination)(nil))
	mapper.CreateMap((*NestedSource)(nil), (*NestedDestination)(nil))
	mapper.Init()

	mapper.Map(src, dest)
}

type Source struct {
	ID      int `mapper:"Id"`
	Name    string
	Weight  float32
	Marks   []int32
	Address []NestedSource
}

type NestedSource struct {
	State string
	Index int
}

type Destination struct {
	Id      int
	Name    string
	Weight  float32
	Marks   []int32
	Address []NestedDestination
}

type NestedDestination struct {
	State string
	Index int
}
