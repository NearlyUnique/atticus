package atticus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"text/template"
)

// ApplyBodyTemplate as configured using the data
func ApplyBodyTemplate(template interface{}, data *TemplateData) ([]byte, error) {
	result := make(map[string]interface{})

	switch t := template.(type) {
	case map[string]interface{}:
		walkMap(t, mapWriter(result), data)
	}
	fmt.Printf("%v\n", reflect.TypeOf(template))
	return json.Marshal(result)
}

type valueWriter func(key string, value interface{})

func mapWriter(m map[string]interface{}) valueWriter {
	return func(key string, value interface{}) {
		m[key] = value
	}
}
func sliceWriter(s *[]interface{}) valueWriter {
	return func(k string, value interface{}) {
		fmt.Printf("sliceWriter: %v , %v", k, value)
		*s = append(*s, value)
	}
}

func walkMap(m map[string]interface{}, result valueWriter, data *TemplateData) {
	for k, v := range m {
		switch value := v.(type) {

		case string:
			result(k, applyStringTemplate(value, data))

		case map[string]interface{}:
			sub := make(map[string]interface{})
			result(k, sub)
			walkMap(value, mapWriter(sub), data)

		case []interface{}:
			var sub []interface{}
			sw := sliceWriter(&sub)

			msub := make(map[string]interface{})

			for _, msub["_"] = range value {
				walkMap(msub, sw, data)
			}

			result(k, sub)

		default:
			result(k, value)
		}
	}
}

func applyStringTemplate(input string, data *TemplateData) string {
	t, err := template.New("test").Parse(input)
	if err != nil {
		return fmt.Sprintf("TEMPLATE_ERROR:%s => (%s)", input, err.Error())
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, data)
	if err != nil {
		return fmt.Sprintf("TEMPLATE_ERROR:%s => (%s)", input, err.Error())
	}
	s := buf.String()
	return s
}
