package unit

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/b1101/systemgo/lib/test"
)

var TEST_INTS = []int{1, 2, 3}
var DEFAULT_UNIT = `[Unit]
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

func TestParse(t *testing.T) {
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
			//{definition{}, false,
			strings.NewReader(DEFAULT_UNIT + `
[Test]
Ints=1 2 3
Bool=yes`),
		},
	}

	for _, c := range cases {
		if err := parseDefinition(c.contents, c.def); err != nil && c.correct {
			t.Errorf(test.ErrorIn, "parseDefinition", err)
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
						t.Errorf(test.Mismatch, option, option.Name)
					}
				case reflect.Bool:
					if option.Bool() != true {
						t.Errorf(test.MismatchIn, option.Name, option.Bool(), true)
					}
				case reflect.Slice:
					if slice, ok := option.Interface().([]string); ok {
						if !reflect.DeepEqual(slice, []string{option.Name}) {
							t.Errorf(test.MismatchIn, option.Name, slice, []string{option.Name})
						}
					} else if slice, ok := option.Interface().([]int); ok {
						if !reflect.DeepEqual(slice, TEST_INTS) {
							t.Errorf(test.MismatchIn, option.Name, slice, TEST_INTS)
						}
					}
				}
			}
		}
	}
}
