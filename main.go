package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/raphaeldichler/zeus/internal/assert"
)

type table struct {
	headers   []string
	numCol    int
	colWidths []int
}

func newTable(entries []*object) *table {
	numCol := entries[0].numEntries()
	colWidths := make([]int, numCol)
	headers := make([]string, numCol)

	for objName := range entries[0].entries {
		headers[0] = objName
	}

	for _, e := range entries {
		i := 0
		for objName, objValue := range e.entries {
			longer := max(len(objName), len(objValue))
			longer += 4 // some extra padding
			colWidths[i] = max(colWidths[i], longer)
			i += 1
		}
	}

	return &table{}
}

func (t *table) writeHeader(w *strings.Builder) {
	for i, header := range t.headers {
		w.WriteString(fmt.Sprintf("%s", header))
		pad := strings.Repeat(" ", t.colWidths[i]-len(header))
		w.WriteString(pad)
	}
	w.WriteString("\n")

	for i, header := range t.headers {
		w.WriteString(strings.Repeat("-", len(header)))
		pad := strings.Repeat(" ", t.colWidths[i]-len(header))
		w.WriteString(pad)
	}
}

type object struct {
	entries map[string]string
}

func (o *object) numEntries() int {
	return len(o.entries)
}

func newObject(entries map[string]string) *object {
	return &object{
		entries: entries,
	}
}
func (o *object) string() string {
	var out string
	for k, v := range o.entries {
		out += fmt.Sprintf("%s: %s\n", k, v)
	}
	return out
}

func toTableObject(obj any) *object {
	switch v := obj.(type) {
	case map[string]any:
		m := make(map[string]string)
		for k, val := range v {
			m[k] = fmt.Sprintf("%v", val)
		}
		return newObject(m)

	default:
		assert.Unreachable("table object cannot have recursive structure")
	}

	return nil
}

func main() {
	jsonStr := `{
    "name": "Alice",
    "age": 30,
    "isActive": true,
    "address": {
      "city": "Wonderland",
      "zipcode": "12345"
    },
    "skills": ["Go", "Python", "Docker"],
    "projects": [
      {
        "name": "Project1",
        "language": "Go"
      },
      {
        "name": "Project2",
        "language": "Python"
      }
    ]
  }`
	jsonStr = `{
    "array": [
      {
        "city": "Wonderland",
        "zipcode": "12345"
      },
      {
        "city": "Wonderland",
        "zipcode": "12345"
      }
    ]
  }`

	// invalid json: "[1, 2, 3]"

	var obj map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		panic(err)
	}

	// Walk through the JSON
	selectType(obj)
}

func walkTable(table any) {
	switch v := table.(type) {
	case []any:
		entries := make([]*object, len(v))
		for _, val := range v {
			obj := toTableObject(val)
			entries = append(entries, obj)
		}

		t := newTable(entries)
		w := strings.Builder{}
		t.writeHeader(&w)

		fmt.Println(w.String())

	default:
		fmt.Println("cannot be anything else than array", v)
	}
}

func selectType(value any) {
	switch v := value.(type) {
	case map[string]any:
		fmt.Println("object")
		for k, val := range v {
			fmt.Printf("Key: %s, Value: %v\n", k, val)
			selectType(val)
		}
	case []any:
		fmt.Println("array")
		walkTable(v)
	default:
		fmt.Println("primitive")
		fmt.Printf("Value: %v\n", v)
	}

}

// walk recursively prints keys and values
func walk(prefix string, value any) {
	switch v := value.(type) {
	case map[string]any:

		fmt.Println("object", v)
		for k, val := range v {

			fmt.Printf("Key: %s, Value: %v\n", k, val)
		}

	default:
		fmt.Println("cannot be anything else than object", v)
	}

	/*
		switch v := value.(type) {

		case map[string]any:
			fmt.Println("object")
			for k, val := range v {
				fullKey := k
				if prefix != "" {
					fullKey = prefix + "." + k
				}
				walk(fullKey, val)
			}
		case []any:
			fmt.Println("array")
			for i, val := range v {
				walk(fmt.Sprintf("%s[%d]", prefix, i), val)
			}
		default:
			fmt.Println("primitive")
			fmt.Printf("Key: %s, Value: %v\n", prefix, v)
		}*/
}
