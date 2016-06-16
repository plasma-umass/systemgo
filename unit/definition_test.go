package unit

import (
	"reflect"
	"strings"
	"testing"
)

func TestDefinition(t *testing.T) {
	var definition Definition

	contents := strings.NewReader(`[Unit]
Description=Description
Documentation=Documentation

Wants=Wants
Requires=Requires
Conflicts=Conflicts
Before=Before
After=After

[Install]
WantedBy=WantedBy
RequiredBy=RequiredBy`)

	if err := ParseDefinition(contents, &definition); err != nil {
		t.Errorf("Error parsing definition: %s", err)
	}

	def := reflect.ValueOf(&definition).Elem()
	for i := 0; i < def.NumField(); i++ {

		section := def.Field(i)
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
					t.Errorf("%s != %s", option, option.Name)
				}
			case reflect.SliceOf(reflect.TypeOf(reflect.String)).Kind():
				slice := option.Interface().([]string)

				if !reflect.DeepEqual(slice, []string{option.Name}) {
					t.Errorf("%s != %s", slice, []string{option.Name})
				}
			}
		}
	}
}
