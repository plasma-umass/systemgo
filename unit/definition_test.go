package unit_test

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/rvolosatovs/systemgo/unit"
	"github.com/stretchr/testify/assert"
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

func TestParseDefinition(t *testing.T) {
	cases := []struct {
		def      interface{}
		correct  bool
		contents io.Reader
	}{
		{&unit.Definition{}, true,
			strings.NewReader(DEFAULT_UNIT),
		},
		{&unit.Definition{}, false,
			strings.NewReader(DEFAULT_UNIT + `
Wrong=Field
Test=should fail`),
		},
		{&struct {
			unit.Definition
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
			unit.Definition
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
			unit.Definition
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
		err := unit.ParseDefinition(c.contents, c.def)
		if !c.correct {
			assert.Error(t, err, "ParseDefinition")
			continue
		}
		if !assert.NoError(t, err, "ParseDefinition") {
			continue
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
					assert.Equal(t, option.String(), option.Name, "string")

					m := methodByName(defVal, option.Name).(func() string)
					assert.Equal(t, m(), option.Name, "string getter")

				case reflect.Bool:
					assert.Equal(t, option.Bool(), DEFAULT_BOOL, "bool")

					// Workaround for the non-existent bool getter
					if defVal.MethodByName(option.Name).IsValid() {
						m := methodByName(defVal, option.Name).(func() bool)
						assert.Equal(t, m(), DEFAULT_BOOL, "bool getter")
					}
				case reflect.Slice:
					if slice, ok := interfaceOf(option.Value).([]string); ok {
						expect := []string{option.Name}
						assert.Equal(t, slice, expect, "[]string")

						m := methodByName(defVal, option.Name).(func() []string)
						assert.Equal(t, m(), expect, "[]string getter")

					} else if slice, ok := interfaceOf(option.Value).([]int); ok {
						assert.Equal(t, slice, DEFAULT_INTS, "[]int")

						// Workaround for the non-existent []int getter
						if defVal.MethodByName(option.Name).IsValid() {
							m := methodByName(defVal, option.Name).(func() []string)
							assert.Equal(t, m(), DEFAULT_INTS, "[]int getter")
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
