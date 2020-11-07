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
		Address: []*NestedSource{
			{
				State: "RYA",
				Index: map[int]SourceMapType{
					1: {Name: "Name1"},
				},
			},
			{
				State: "MOW",
				Index: map[int]SourceMapType{
					2: {Name: "Name2"},
				},
			},
		},
	}

	dest := &Destination{}

	testMapper := mapper.NewMapper()
	testMapper.CreateMap((*Source)(nil), (*Destination)(nil))
	testMapper.CreateMap((*NestedSource)(nil), (*NestedDestination)(nil))
	testMapper.Init()

	testMapper.Map(src, dest)
}

type Source struct {
	ID      int `mapper:"Id"`
	Name    string
	Weight  float32
	Marks   []int32
	Address []*NestedSource
}

type NestedSource struct {
	State string
	Index map[int]SourceMapType
}

type Destination struct {
	Id        int
	FirstName string `mapper:"Name"`
	Weight    float32
	Marks     []int32
	Address   []*NestedDestination
}

type NestedDestination struct {
	State string
	Index map[int]DestinationMapType
}

type SourceMapType struct {
	Name string
}

type DestinationMapType struct {
	Name string
}
