package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/d4l3k/messagediff"
	"os"
	"reflect"
	"sort"
	"strings"
)

func main() {
	os.Exit(rMain())
}

func rMain() int {
	var err error

	var dumps [2]map[string]UefiVar

	var offset int
	var tFile []byte

	var interestingVar = flag.String("f", "", "only results from this variable will be printed")

	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Println("invalid number of arguments")
		return 1
	}

	//load diffs from file into maps
	for i := 0; i < 2; i++ {
		tFile, err = os.ReadFile(flag.Arg(i))
		if err != nil {
			fmt.Printf("error opening file %s: %v\n", flag.Arg(i), err)
			return 1
		}

		dumps[i] = make(map[string]UefiVar)
		offset, err = jsonparser.ArrayEach(tFile, func(value []byte, dataType jsonparser.ValueType, offset int, er error) {
			defer func() {
				if er != nil {
					os.Exit(1)
				}
			}()
			if er != nil {
				fmt.Printf("error parsing json at offset %d: %v\n", offset, er)
				return
			}

			var tVar UefiVar

			er = json.Unmarshal(value, &tVar)
			if er != nil {
				fmt.Printf("error unmarshalling var at %d: %v\n", offset, er)
				return
			}

			if *interestingVar != "" {
				if tVar.Name == *interestingVar {
					dumps[i][tVar.Name] = tVar
				}
			} else { //no filter
				dumps[i][tVar.Name] = tVar
			}
		})
		if err != nil {
			fmt.Printf("error parsing %s at offset %d: %v", flag.Arg(i), offset, err)
			return 1
		}
	}

	d, _ := PrettyDiffCompare(dumps[0], dumps[1])
	fmt.Print(d)

	return 0
}

func PrettyDiffCompare(a, b interface{}) (string, bool) {
	d, equal := messagediff.DeepDiff(a, b)
	var dstr []string
	for path, added := range d.Added {
		dstr = append(dstr, fmt.Sprintf("added: %s = %#v\n", path.String(), added))
	}
	for path, removed := range d.Removed {
		dstr = append(dstr, fmt.Sprintf("removed: %s = %#v\n", path.String(), removed))
	}
	for path, modified := range d.Modified {
		dstr = append(dstr, fmt.Sprintf("modified: %s = %#v | a[%#v, %#v]b\n", path.String(), modified, valueFromPath(a, path), valueFromPath(b, path)))
	}
	sort.Strings(dstr)
	return strings.Join(dstr, ""), equal
}

//var ErrNoValue = errors.New("no value found for key")

func valueFromPath(base interface{}, path *messagediff.Path) interface{} {
	var v reflect.Value = reflect.ValueOf(base)

	for _, p := range *path {
		v = valueFromKey(v, p)
	}

	return v
}

// no checks since it should always be fed by diff output, which is assumed always valid
func valueFromKey(base reflect.Value, key messagediff.PathNode) reflect.Value {
	if base.IsZero() {
		panic("zero value")
	}
	switch base.Kind() {
	case reflect.Array, reflect.Slice:
		return base.Index(int(key.(messagediff.SliceIndex)))
	case reflect.Map:
		return base.MapIndex(reflect.ValueOf(key.(messagediff.MapKey).Key))
	case reflect.Struct:
		return base.FieldByName(string(key.(messagediff.StructField)))
	case reflect.Ptr:
		return valueFromKey(base.Elem(), key)
	default:
		panic(fmt.Sprintf("unsupported type %v", base.Kind()))
	}
}

type UefiVar struct {
	Name       string `json:"name"`
	VendorGuid string `json:"vendor_guid"`

	Attributes struct {
		NonVolatile                       bool `json:"non_volatile"`
		BootserviceAccess                 bool `json:"bootservice_access"`
		RuntimeAccess                     bool `json:"runtime_access"`
		HardwareErrorRecord               bool `json:"hardware_error_record"`
		AuthenticatedWriteAccess          bool `json:"authenticated_write_access"`
		TimeBasedAuthenticatedWriteAccess bool `json:"time_based_authenticated_write_access"`
		EnhancedAuthenticatedAccess       bool `json:"enhanced_authenticated_access"`
	} `json:"attributes"`

	DataLen int    `json:"data_len"`
	Data    []byte `json:"data"`
}
