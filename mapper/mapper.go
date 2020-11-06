package mapper

import (
	"github.com/golang/glog"

	"reflect"
)

func NewMapper() *Mapper {
	return &Mapper{
		profiles: make(map[reflect.Type]map[string]string),
		maps:     make(map[reflect.Type]reflect.Type),
	}
}

type Mapper struct {
	profiles map[reflect.Type]map[string]string
	maps     map[reflect.Type]reflect.Type
}

// CreateMap func creates new spec for types mapping.
func (o *Mapper) CreateMap(src interface{}, dest interface{}) *Mapper {
	o.maps[reflect.TypeOf(src).Elem()] = reflect.TypeOf(dest).Elem()
	// TODO remove workaround with slices mapping to MapSlices() extension
	o.maps[reflect.SliceOf(reflect.TypeOf(src).Elem())] = reflect.SliceOf(reflect.TypeOf(dest).Elem())

	return o
}

// TODO reverse maps extension.
func (o *Mapper) Reverse() *Mapper {
	return o
}

// TODO slices mapping extension.
func (o *Mapper) MapSlices() *Mapper {
	return o
}

func (o *Mapper) Init() {
	for src, dest := range o.maps {
		srcMeta := o.getTypeMeta(src)
		if srcMeta != nil {
			o.profiles[src] = srcMeta
		}

		destMeta := o.getTypeMeta(dest)
		if destMeta != nil {
			o.profiles[dest] = destMeta
		}
	}
}

// getTypeMeta func fetches struct keys and mapper tags, e.g. map[key]tag.
func (o *Mapper) getTypeMeta(val reflect.Type) map[string]string {
	if val.Kind() != reflect.Struct {
		return nil
	}

	fieldsCount := val.NumField()

	res := make(map[string]string, fieldsCount)

	for i := 0; i < fieldsCount; i++ {
		typeField := val.Field(i)
		res[typeField.Name] = typeField.Tag.Get("mapper")
	}

	return res
}

func (o *Mapper) Map(src interface{}, dest interface{}) {
	srcVal := reflect.ValueOf(src)

	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr {
		glog.Errorf("provided destination has invalid kind: expected reflect.Ptr, got: %s", destVal.Kind().String())
		return
	}

	if srcVal.Kind() != destVal.Elem().Kind() {
		glog.Errorf("unable to map %s into %s", srcVal.Kind().String(), destVal.Elem().Kind().String())
		return
	}

	switch srcVal.Kind() {
	case reflect.Struct:
		o.mapStructs(srcVal, destVal.Elem())
	case reflect.Slice:
		o.mapSlices(srcVal, destVal.Elem())
	case reflect.Map:
		glog.Infoln("not supported yet")
		return
	case reflect.Ptr:
		glog.Infoln("not supported yet")
		return
	default:
		return
	}
}

// mapStructs func perform structs casts.
func (o *Mapper) mapStructs(src reflect.Value, dest reflect.Value) {
	// Get structs types
	// if types are equal set dest slice
	if src.Type() == dest.Type() {
		dest.Set(src)
		return
	}

	// Get structs types
	// if types were not registered - abort
	if o.maps[src.Type()] != dest.Type() {
		glog.Errorf("no maps specified for types %s and %s", src.Type().String(), dest.Type().String())
		return
	}

	// get keys and tags maps foe src and dest structs
	srcMap, ok := o.profiles[src.Type()]
	if !ok {
		glog.Errorf("no profile specfied for %s", dest.Type().String())
		return
	}

	destMap, ok := o.profiles[dest.Type()]
	if !ok {
		glog.Errorf("no profile specfied for %s", dest.Type().String())
		return
	}

	// iterate over struct fields and map values
	for key, tag := range srcMap {
		// TODO resolve dest tagging problem, now only src tags has value
		if _, ok := destMap[key]; !ok {
			if _, ok := destMap[tag]; !ok {
				continue
			}
		}

		srcVal := src.FieldByName(key)
		destVal := dest.FieldByName(key)

		o.processValues(srcVal, destVal)
	}
}

// mapSlices func perform slices casts.
func (o *Mapper) mapSlices(src reflect.Value, dest reflect.Value) {
	// Get slice type
	// if types are equal set dest slice
	if src.Type() == dest.Type() {
		dest.Set(src)
		return
	}

	// Get slice type
	// if types were not registered - abort
	if o.maps[src.Type()] != dest.Type() {
		glog.Errorf("no maps specified for types %s and %s", src.Type().String(), dest.Type().String())
		return
	}

	// Make dest slice
	dest.Set(reflect.MakeSlice(dest.Type(), src.Len(), src.Cap()))

	// Get each element of slice
	// check its kind and try to map/set dest
	for i := 0; i < src.Len(); i++ {
		srcVal := src.Index(i)
		destVal := dest.Index(i)

		o.processValues(srcVal, destVal)
	}
}

// processValues func resolve src and dest values kind
// and either recursively calls mapping functions, or sets dest value.
func (o *Mapper) processValues(src reflect.Value, dest reflect.Value) {
	// get provided values' kinds
	srcKind := src.Kind()
	destKind := dest.Kind()

	// check if kinds are equal
	if srcKind != destKind {
		// TODO dynamic cast, m.b. with mapper extensions
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
		glog.Infoln("not supported yet")
		return
	case reflect.Ptr:
		glog.Infoln("not supported yet")
		return
	default:
		dest.Set(src)
	}
}
