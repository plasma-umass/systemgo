package unit

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/b1101/systemgo/lib/test"
)

var DEFAULT_INTS = []int{1, 2, 3}

const DEFAULT_BOOL = true
const DEFAULT_UNIT = `[Unit]
Description=Description
Documentation=Documentation

Wants=Wants
Requires=Requires
Conflicts=Conflicts
Before=Before
After=After

[Install]
WantedBy=WantedBy
RequiredBy=RequiredBy`

func TestDefinition(t *testing.T) {
	cases := []struct {
		def      interface{}
		correct  bool
		contents io.Reader
	}{
		{&definition{}, true,
			strings.NewReader(DEFAULT_UNIT),
		},
		{&definition{}, false,
			strings.NewReader(DEFAULT_UNIT + `
Wrong=Field
Test=should fail`),
		},
		{&struct {
			definition
			Test struct {
				Ints []int
				Bool bool
			}
		}{}, true,
			strings.NewReader(DEFAULT_UNIT + `
[Test]
Ints=1 2 3
Bool=yes`),
		},
		{&struct {
			definition
			Test struct {
				Ints []int
				Bool bool
			}
		}{}, false,
			strings.NewReader(DEFAULT_UNIT + `
[Test]
Ints=1 2 3
Bool=foo`),
		},
		{&struct {
			definition
			Test struct {
				Ints []int
				Bool bool
			}
		}{}, false,
			strings.NewReader(DEFAULT_UNIT + `
[Test]
Ints=a b 3
Bool=foo`),
		},
	}

	for _, c := range cases {
		if err := parseDefinition(c.contents, c.def); err != nil {
			if c.correct {
				t.Errorf(test.ErrorIn, "parseDefinition", err)
			}
			continue
		} else if !c.correct && err == nil {
			t.Errorf(test.Nil, "err")
		}

		defVal := reflect.ValueOf(c.def).Elem()
		for i := 0; i < defVal.NumField(); i++ {

			section := defVal.Field(i)
			sectionType := section.Type()

			for j := 0; j < section.NumField(); j++ {
				option := struct {
					reflect.Value
					Name string
				}{
					section.Field(j),
					sectionType.Field(j).Name,
				}

				switch option.Kind() {
				case reflect.String:
					if option.String() != option.Name {
						t.Errorf(test.MismatchIn, option.Name, option, option.Name)
					}
					if m := methodByName(defVal, option.Name).(func() string); m() != option.Name {
						t.Errorf(test.MismatchIn, option.Name+"()", m(), option.Name)
					}
				case reflect.Bool:
					if option.Bool() != DEFAULT_BOOL {
						t.Errorf(test.MismatchInVal, option.Name, option.Bool(), DEFAULT_BOOL)
					}
					// Workaround for the non-existent bool getter
					if defVal.MethodByName(option.Name).IsValid() {
						if m := methodByName(defVal, option.Name).(func() bool); m() != DEFAULT_BOOL {
							t.Errorf(test.MismatchInVal, option.Name+"()", m(), DEFAULT_BOOL)
						}
					}
				case reflect.Slice:
					if slice, ok := interfaceOf(option.Value).([]string); ok {
						expect := []string{option.Name}

						if !reflect.DeepEqual(slice, expect) {
							t.Errorf(test.MismatchIn, option.Name, slice, expect)
						}
						if m := methodByName(defVal, option.Name).(func() []string); !reflect.DeepEqual(m(), expect) {
							t.Errorf(test.MismatchIn, option.Name+"()", m(), expect)
						}
					} else if slice, ok := interfaceOf(option.Value).([]int); ok {
						if !reflect.DeepEqual(slice, DEFAULT_INTS) {
							t.Errorf(test.MismatchInVal, option.Name, slice, DEFAULT_INTS)
						}
						// Workaround for the non-existent []int getter
						if defVal.MethodByName(option.Name).IsValid() {
							if m := methodByName(defVal, option.Name).(func() []string); !reflect.DeepEqual(m(), DEFAULT_INTS) {
								t.Errorf(test.MismatchIn, option.Name+"()", m(), DEFAULT_INTS)
							}
						}
					}
				}
			}
		}
	}
}

func interfaceOf(val reflect.Value) interface{} {
	return val.Interface()
}

func methodByName(val reflect.Value, name string) interface{} {
	return interfaceOf(val.MethodByName(name))
}
