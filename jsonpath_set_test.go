package jsonpath

import (
	"encoding/json"
	"testing"
)

type JsonpathSetCase struct {
	name        string
	expr        string
	data        string
	change      interface{}
	isErrorCase bool
}

func SetCases() []JsonpathSetCase {
	return []JsonpathSetCase{
		{
			name:   "multi-level virtual elements with empty",
			expr:   "$.a.b.c.d.e",
			data:   "{}",
			change: nil,
		},
		{
			name:   "multi-level virtual elements with data",
			expr:   "$.a.b.c.d.e",
			data:   `{"a":{"b":{"c": {"x": "y"}}}}`,
			change: nil,
		},
		{
			name:   "multi-level virtual elements with data over expr",
			expr:   "$.a.b.c.d.e",
			data:   `{"a":{"b":{"c":{"d":{"e":{"f":"some chars"},"x":"y"}}}}}`,
			change: nil,
		},
		{
			name:   "single index in empty array",
			expr:   "$[0]",
			data:   `[]`,
			change: false,
		},
		{
			name:   "single index in array",
			expr:   "$[0]",
			data:   `[0,1,2,3,4,5,6]`,
			change: false,
		},
		{
			name:   "range indexes in array",
			expr:   "$[1:3]",
			data:   `[0,1,2,3,4,5,6]`,
			change: false,
		},
		{
			name:   "range indexes in empty array",
			expr:   "$[1:3]",
			data:   `[]`,
			change: false,
		},
	}
}

func TestSetFunction(t *testing.T) {
	cases := SetCases()
	//cases := SetCases()[3:4]
	for _, c := range cases {
		j, err := New(c.name, c.expr)
		if err != nil {
			t.Fatalf("cannot parse jsonpath")
		}
		jsonObj := ConvertToJsonObj(c.data)
		j.InitData(jsonObj)
		err = j.Set(c.change)
		if err != nil {
			t.Errorf(err.Error())
		} else {
			marshal, err := json.Marshal(j.Data())
			if err != nil {
				t.Errorf("json marshal error: %s", err)
			}
			t.Logf("success: %s", marshal)
		}
	}
}
