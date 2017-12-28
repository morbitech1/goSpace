package container

import (
	"crypto/sha256"
	"fmt"
	"github.com/pspaces/gospace/function"
	"reflect"
	"sort"
	"strings"
)

// Signature represents structure for a signature of values.
type Signature interface{}

// NewSignature creates a hashed signature by recursively traversing the value val.
// NewSignature expects a recursion depth of rd if it encounters structs, maps or slices.
func NewSignature(rd uint, val interface{}) (s Signature) {
	var sign interface{}

	halg := sha256.New()

	halg.Write([]byte(fmt.Sprintf("%+v", reflect.TypeOf(val))))

	switch kind := reflect.ValueOf(val).Kind(); kind {
	case reflect.Func:
		halg.Write([]byte(fmt.Sprintf("%v%v%v", reflect.ValueOf(val).Pointer(), function.Name(val), function.Signature(val))))
	case reflect.Array, reflect.Slice:
		if rd >= 0 {
			params := val.([]interface{})
			for _, param := range params {
				subsign := NewSignature(rd-1, param)
				halg.Write([]byte(fmt.Sprintf("%+v", reflect.TypeOf(param))))
				halg.Write([]byte(subsign.(string)))
			}
		} else {
			halg.Write([]byte(fmt.Sprintf("%+v", val)))
		}
	case reflect.Map:
		if rd >= 0 {
			rmap := reflect.ValueOf(val)
			keys := rmap.MapKeys()

			shkeys := make([]string, 0, len(keys))
			hmap := make(map[string]string)
			for _, key := range keys {
				keySign := NewSignature(rd-1, key.Interface())
				keyTypeSign := NewSignature(rd-1, fmt.Sprintf("%+v", reflect.TypeOf(key.Interface())))
				val := rmap.MapIndex(key)
				valSign := NewSignature(rd-1, val.Interface())
				valTypeSign := NewSignature(rd-1, fmt.Sprintf("%+v", reflect.TypeOf(val.Interface())))
				hkey := fmt.Sprintf("%+v %+v", keySign, keyTypeSign)
				hval := fmt.Sprintf("%+v %+v", valSign, valTypeSign)
				hmap[hkey] = hval
				shkeys = append(shkeys, hkey)
			}
			sort.Strings(shkeys)

			for _, hkey := range shkeys {
				ksts := strings.Split(hkey, " ")
				for _, subsign := range ksts {
					halg.Write([]byte(fmt.Sprintf("%+v", subsign)))
				}

				vsts := strings.Split(hmap[hkey], " ")
				for _, subsign := range vsts {
					halg.Write([]byte(fmt.Sprintf("%+v", subsign)))
				}
			}
		} else {
			halg.Write([]byte(fmt.Sprintf("%+v", val)))
		}
	case reflect.Struct:
		if rd >= 0 {
			rstruct := reflect.ValueOf(val)
			cnt := rstruct.NumField()

			// In `reflect`, Field() also behaves non-deterministically. It sucks.
			// One can not rely on uniqueness of field names, content or type,
			// but forced to sort by signature of value and type.
			// Otherwise, this behaviour causes unintended side-effects.
			// See also: https://golang.org/src/reflect/type.go?s=7347:7852#L199
			svtkeys := make([]string, 0, cnt)
			for i := 0; i < cnt; i++ {
				field := rstruct.Field(i)
				valSign := NewSignature(rd-1, field.Interface())
				typeSign := NewSignature(rd-1, fmt.Sprintf("%+v", field.Type()))
				svtkey := fmt.Sprintf("%+v %+v", valSign, typeSign)
				svtkeys = append(svtkeys, svtkey)
			}
			sort.Strings(svtkeys)

			for _, svtkey := range svtkeys {
				vsts := strings.Split(svtkey, " ")
				for _, subsign := range vsts {
					halg.Write([]byte(fmt.Sprintf("%+v", subsign)))
				}
			}
		} else {
			halg.Write([]byte(fmt.Sprintf("%+v", val)))
		}
	case reflect.Ptr:
		rptr := reflect.ValueOf(val)
		rval := reflect.Indirect(rptr).Interface()
		rkind := reflect.ValueOf(rval).Kind()
		if rkind == reflect.Array || rkind == reflect.Slice || rkind == reflect.Map || rkind == reflect.Struct {
			subsign := NewSignature(rd, rval)
			halg.Write([]byte(fmt.Sprintf("%+v", subsign)))
		} else {
			halg.Write([]byte(fmt.Sprintf("%+v", rptr.Interface())))
		}
	default:
		halg.Write([]byte(fmt.Sprintf("%+v", val)))
	}

	sign = fmt.Sprintf("%x", halg.Sum(nil))

	s = sign

	return s
}
