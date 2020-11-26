package mapper

import (
	"flag"
	"fmt"
	"reflect"

	"github.com/golang/glog"
)

const (
	SrcKeyIndex = iota
	DestKeyIndex
)

// NewMapper func returns new uninitialised Mapper.
// New maps should be created with CreateMap func.
// After creating maps call Init to initialise Mapper.
func NewMapper() *Mapper {
	return &Mapper{
		isInitialised: false,
		profiles:      make(map[string][][2]string),
		profilesOpts:  make(map[string]*profileOptions),
	}
}

// Mapper struct contains maps for registered pairs of types
// and profiles for resolving struct fields conversions.
type Mapper struct {
	isInitialised bool                            // checks if Mapper was initialised before usage
	profiles      map[string][][2]string          // map of struct fields: ["srcType_destType"][]["src_key", "dest_key"]
	profilesOpts  map[string]*profileOptions      // options for profiles, such as reverse mapping, etc
	maps          []map[reflect.Type]reflect.Type // pairs of types to map
}

// typeMeta struct contains meta info about struct fields
// used to resolve conventions between fields names and tags.
type typeMeta struct {
	keysToTags map[string]string
	tagsToKeys map[string]string
}

// profileOptions struct contains additional data for structs mapping.
type profileOptions struct {
	reverseMap bool
}

// getProfileKey converts src and dest types in string key representation.
func getProfileKey(srcType reflect.Type, destType reflect.Type) string {
	return fmt.Sprintf("%s_%s", srcType.Name(), destType.Name())
}

// CreateMap func creates new spec for types mapping.
// CreateMap should be called ONLY before Init function call.
// Provided map can be reversed with chained Reverse function:
//	CreateMap((*Source)(nil), (*Destination)(nil)).Reverse()
// You can create conversion between slices with MapSlices func
//	CreateMap((*Source)(nil), (*Destination)(nil)).MapSlices().
func (o *Mapper) CreateMap(src interface{}, dest interface{}) *Mapper {
	typesMap := make(map[reflect.Type]reflect.Type)
	typesMap[reflect.TypeOf(src).Elem()] = reflect.TypeOf(dest).Elem()

	o.maps = append(o.maps, typesMap)

	return o
}

// Reverse func reverts last created map.
func (o *Mapper) Reverse() *Mapper {
	for srcType, destType := range o.maps[len(o.maps)-1] {
		o.profilesOpts[getProfileKey(srcType, destType)].reverseMap = true
	}

	return o
}

// Init func fills profiles from provided types maps.
func (o *Mapper) Init() {
	// parse logger flags
	flag.Parse()

	for _, typesMap := range o.maps {
		for srcType, destType := range typesMap {
			// check for provided types kind.
			// if not struct - skip.
			if srcType.Kind() != reflect.Struct {
				glog.Errorf("expected reflect.Struct kind for type %s, but got %s", srcType.String(), srcType.Kind().String())
				continue
			}

			if destType.Kind() != reflect.Struct {
				glog.Errorf("expected reflect.Struct kind for type %s, but got %s", destType.String(), destType.Kind().String())
				continue
			}

			// if a reverse flag for given types exists add reverse map
			if options, ok := o.profilesOpts[getProfileKey(srcType, destType)]; ok &&
				options.reverseMap {
				typesMap[destType] = srcType
			}

			// profile is slice of src and dest structs fields names
			var profile [][2]string

			// get types metadata
			srcMeta := o.getTypeMeta(srcType)
			destMeta := o.getTypeMeta(destType)

			for srcKey, srcTag := range srcMeta.keysToTags {
				// case src key equals dest key
				if _, ok := destMeta.keysToTags[srcKey]; ok {
					profile = append(profile, [2]string{srcKey, srcKey})
					continue
				}

				// case src key equals dest tag
				if destKey, ok := destMeta.tagsToKeys[srcKey]; ok {
					profile = append(profile, [2]string{srcKey, destKey})
					continue
				}

				// case src tag equals dest key
				if _, ok := destMeta.keysToTags[srcTag]; ok {
					profile = append(profile, [2]string{srcKey, srcTag})
					continue
				}

				// case src tag equals dest tag
				if destKey, ok := destMeta.tagsToKeys[srcTag]; ok {
					profile = append(profile, [2]string{srcKey, destKey})
					continue
				}
			}

			// save profile with unique srcKey for provided types
			o.profiles[getProfileKey(srcType, destType)] = profile
		}
	}

	o.isInitialised = true
}

// getTypeMeta func fetches struct fields keysToTags, types and Mapper tags.
func (o *Mapper) getTypeMeta(val reflect.Type) typeMeta {
	fieldsNum := val.NumField()

	keysToTags := make(map[string]string)
	tagsToKeys := make(map[string]string)

	for i := 0; i < fieldsNum; i++ {
		field := val.Field(i)
		fieldName := field.Name
		fieldTag := field.Tag.Get("mapper")

		keysToTags[fieldName] = fieldTag
		tagsToKeys[fieldTag] = fieldName
	}

	return typeMeta{
		keysToTags: keysToTags,
		tagsToKeys: tagsToKeys,
	}
}

// Map func checks for initialised Mapper and starts types mapping process.
// Should be called ONLY after Init function call.
func (o *Mapper) Map(src interface{}, dest interface{}) {
	// stop mapping if Mapper was not initialised
	if !o.isInitialised {
		glog.Error("uninitialised Mapper usage is permitted. You should call Init() func before Map() calling")
		return
	}

	// check if provided dest has pointer kind.
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr {
		glog.Errorf("provided destination has invalid kind: expected reflect.Ptr, got: %s", destVal.Kind().String())
		return
	}

	// start values processing
	o.processValues(reflect.ValueOf(src), destVal.Elem())
}

// processValues func resolve src and dest values kind
// and either recursively calls mapping functions, or sets dest value.
func (o *Mapper) processValues(src reflect.Value, dest reflect.Value) {
	// If src of dest is an interface - get underlying type
	if src.Kind() == reflect.Interface {
		src = src.Elem()
	}

	if dest.Kind() == reflect.Interface {
		dest = dest.Elem()
	}

	// get provided values' kinds
	srcKind := src.Kind()
	destKind := dest.Kind()

	// check if kinds are equal
	if srcKind != destKind {
		// TODO dynamic cast, m.b. with Mapper extensions
		return
	}

	// if types are equal set dest value
	if src.Type() == dest.Type() {
		dest.Set(src)
		return
	}

	// resolve kind and choose mapping function
	// or set dest value
	switch src.Kind() {
	case reflect.Struct:
		o.mapStructs(src, dest)
	case reflect.Slice:
		o.mapSlices(src, dest)
	case reflect.Map:
		o.mapMaps(src, dest)
	case reflect.Ptr:
		o.mapPointers(src, dest)
	default:
		dest.Set(src)
	}
}

// mapStructs func perform structs casts.
func (o *Mapper) mapStructs(src reflect.Value, dest reflect.Value) {
	// get values types
	// if types or their slices were not registered - abort
	profile, ok := o.profiles[getProfileKey(src.Type(), dest.Type())]
	if !ok {
		glog.Errorf("no conversion specified for types %s and %s", src.Type().String(), dest.Type().String())
		return
	}

	// iterate over struct fields and map values
	for _, keys := range profile {
		o.processValues(src.FieldByName(keys[SrcKeyIndex]), dest.FieldByName(keys[DestKeyIndex]))
	}
}

// mapSlices func perform slices casts.
func (o *Mapper) mapSlices(src reflect.Value, dest reflect.Value) {
	// Make dest slice
	dest.Set(reflect.MakeSlice(dest.Type(), src.Len(), src.Cap()))

	// Get each element of slice
	// process values mapping
	for i := 0; i < src.Len(); i++ {
		srcVal := src.Index(i)
		destVal := dest.Index(i)

		o.processValues(srcVal, destVal)
	}
}

// mapPointers func perform pointers casts.
func (o *Mapper) mapPointers(src reflect.Value, dest reflect.Value) {
	// create new struct from provided dest type
	val := reflect.New(dest.Type().Elem()).Elem()

	o.processValues(src.Elem(), val)

	// assign address of initialised struct to destination
	dest.Set(val.Addr())
}

// mapMaps func perform maps casts.
func (o *Mapper) mapMaps(src reflect.Value, dest reflect.Value) {
	// Make dest map
	dest.Set(reflect.MakeMapWithSize(dest.Type(), src.Len()))

	// Get each element of map as key-values
	// process keys and values mapping and update dest map
	srcMapIter := src.MapRange()
	destMapIter := dest.MapRange()

	for destMapIter.Next() && srcMapIter.Next() {
		destKey := reflect.New(destMapIter.Key().Type()).Elem()
		destValue := reflect.New(destMapIter.Value().Type()).Elem()

		o.processValues(srcMapIter.Key(), destKey)
		o.processValues(srcMapIter.Value(), destValue)

		dest.SetMapIndex(destKey, destValue)
	}
}
