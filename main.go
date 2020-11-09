package main

import (
	"github.com/alexsem80/go-mapper/mapper"
)

func main() {
	src := Destination{
		Id:     1,
		Name:   "Name",
		Weight: 32.32,
		Marks:  []int32{2, 3, 4},
		Address: []*NestedDestination{
			{
				State: "RYA",
				Index: map[int]*DestinationMapType{
					1: {Name: "Name1"},
				},
			},
			{
				State: "MOW",
				Index: map[int]*DestinationMapType{
					2: {Name: "Name2"},
				},
			},
		},
	}

	dest := &Source{}

	mapper := mapper.NewMapper()
	mapper.CreateMap((*Source)(nil), (*Destination)(nil))
	mapper.CreateMap((*NestedSource)(nil), (*NestedDestination)(nil))
	mapper.CreateMap((*SourceMapType)(nil), (*DestinationMapType)(nil))

	mapper.Init()

	mapper.Map(dest, src)
}

type Source struct {
	ID        int
	FirstName string `mapper:"Name"`
	Weight    float32
	Marks     []int32
	Address   []*NestedSource
}

type NestedSource struct {
	State string
	Index map[int]*SourceMapType
}

type Destination struct {
	Id      int `mapper:"ID"`
	Name    string
	Weight  float32
	Marks   []int32
	Address []*NestedDestination
}

type NestedDestination struct {
	State string
	Index map[int]*DestinationMapType
}

type SourceMapType struct {
	Name string
}

type DestinationMapType struct {
	Name string
}
