package mapper

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var srcType = (*Source)(nil)
var destType = (*Destination)(nil)

var createMapTestData = []struct {
	label             string
	srcType, destType interface{}
	reversed          bool
	wantedMaps        []map[reflect.Type]reflect.Type
}{
	{
		"Case-1: CreateMap Source -> Destination",
		srcType,
		destType,
		false,
		[]map[reflect.Type]reflect.Type{
			{
				reflect.TypeOf(srcType).Elem(): reflect.TypeOf(destType).Elem(),
			},
		},
	},
}

func TestCreateMap(t *testing.T) {
	for _, data := range createMapTestData {
		testMapper := NewMapper()

		testMapper.CreateMap(data.srcType, data.destType)

		assert.True(t, reflect.DeepEqual(testMapper.maps, data.wantedMaps), data.label)
	}
}

type Source struct{}
type Destination struct{}

var mapTestData = []struct {
	label           string
	srcObj, destObj interface{}
}{
	{
		"Case - 1. Map structs",
		SrcStruct{
			ID:     1,
			Name:   "Name",
			Weight: 10.5,
			Marks:  []int32{1, 2, 3},
		},
		DestStruct{
			Id:     1,
			Name:   "Name",
			Weight: 10.5,
			Marks:  []int32{1, 2, 3},
		},
	},
	{
		"Case - 2. Map slices",
		struct {
			slice []SrcStruct
		}{
			[]SrcStruct{
				{
					ID:     1,
					Name:   "Name",
					Weight: 10.5,
					Marks:  []int32{1, 2, 3},
				},
				{
					ID:     2,
					Name:   "Name2",
					Weight: 20.5,
					Marks:  []int32{3, 5, 7},
				},
			},
		},
		struct {
			slice []DestStruct
		}{
			[]DestStruct{
				{
					Id:     1,
					Name:   "Name",
					Weight: 10.5,
					Marks:  []int32{1, 2, 3},
				},
				{
					Id:     2,
					Name:   "Name2",
					Weight: 20.5,
					Marks:  []int32{3, 5, 7},
				},
			},
		},
	},
}

func TestMap(t *testing.T) {
	testMapper := NewMapper()
	testMapper.CreateMap((*SrcStruct)(nil), (*DestStruct)(nil))
	testMapper.Init()

	for _, data := range mapTestData {
		dest := &struct{}{}
		testMapper.Map(data.srcObj, dest)
		assert.True(t, reflect.DeepEqual(reflect.ValueOf(dest).Elem(), data.destObj), data.label)
	}
}

type SrcStruct struct {
	ID     int
	Name   string
	Weight float32
	Marks  []int32
}

type DestStruct struct {
	Id     int `mapper:"ID"`
	Name   string
	Weight float32
	Marks  []int32
}
