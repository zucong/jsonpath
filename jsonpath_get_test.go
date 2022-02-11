package jsonpath

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type JsonpathGetCase struct {
	name        string
	expr        string
	data        string
	expectation string
	isErrorCase bool
}

func LoadGetCases(cases *map[string]JsonpathGetCase) {
	m := *cases
	m["Array slice"] = JsonpathGetCase{
		name:        "Array slice",
		expr:        `$[1:3]`,
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["second","third"]`,
	}
	m["Array slice on exact match"] = JsonpathGetCase{
		name:        "Array slice on exact match",
		expr:        "$[0:5]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["first","second","third","forth","fifth"]`,
	}
	m["Array slice on non overlapping array"] = JsonpathGetCase{
		name:        "Array slice on non overlapping array",
		expr:        "$[7:10]",
		data:        `["first", "second", "third"]`,
		expectation: `[]`,
	}
	m["Array slice on object"] = JsonpathGetCase{
		name:        "Array slice on object",
		expr:        "$[1:3]",
		data:        `{":": 42, "more": "string", "a": 1, "b": 2, "c": 3, "1:3": "nice"}`,
		expectation: `[]`,
	}
	m["Array slice on partially overlapping array"] = JsonpathGetCase{
		name:        "Array slice on partially overlapping array",
		expr:        "$[1:10]",
		data:        `["first", "second", "third"]`,
		expectation: `["second","third"]`,
	}
	m["Array slice with large number for end"] = JsonpathGetCase{
		name:        "Array slice with large number for end",
		expr:        "$[2:113667776004]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["third","forth","fifth"]`,
	}
	m["Array slice with large number for end and negative step"] = JsonpathGetCase{
		name:        "Array slice with large number for end and negative step",
		expr:        "$[2:-113667776004:-1]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["third","second","first"]`,
	}
	m["Array slice with large number for start"] = JsonpathGetCase{
		name:        "Array slice with large number for start",
		expr:        "$[-113667776004:2]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["first","second"]`,
	}
	m["Array slice with large number for start end negative step"] = JsonpathGetCase{
		name:        "Array slice with large number for start end negative step",
		expr:        "$[113667776004:2:-1]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["fifth","forth"]`,
	}
	m["Array slice with negative start and end and range of -1"] = JsonpathGetCase{
		name:        "Array slice with negative start and end and range of -1",
		expr:        "$[-4:-5]",
		data:        `[2, "a", 4, 5, 100, "nice"]`,
		expectation: `[]`,
	}
	m["Array slice with negative start and end and range of 0"] = JsonpathGetCase{
		name:        "Array slice with negative start and end and range of 0",
		expr:        "$[-4:-5]",
		data:        `[2, "a", 4, 5, 100, "nice"]`,
		expectation: `[]`,
	}
	m["Array slice with negative start and end and range of 1"] = JsonpathGetCase{
		name:        "Array slice with negative start and end and range of 1",
		expr:        "$[-4:-3]",
		data:        `[2, "a", 4, 5, 100, "nice"]`,
		expectation: `[4]`,
	}
	m["Array slice with negative start and positive end and range of -1"] = JsonpathGetCase{
		name:        "Array slice with negative start and positive end and range of -1",
		expr:        "$[-4:1]",
		data:        `[2, "a", 4, 5, 100, "nice"]`,
		expectation: `[]`,
	}
	m["Array slice with negative start and positive end and range of 0"] = JsonpathGetCase{
		name:        "Array slice with negative start and positive end and range of 0",
		expr:        "$[-4:2]",
		data:        `[2, "a", 4, 5, 100, "nice"]`,
		expectation: `[]`,
	}
	m["Array slice with negative start and positive end and range of 1"] = JsonpathGetCase{
		name:        "Array slice with negative start and positive end and range of 1",
		expr:        "$[-4:3]",
		data:        `[2, "a", 4, 5, 100, "nice"]`,
		expectation: `[4]`,
	}
	m["Array slice with negative step"] = JsonpathGetCase{
		name:        "Array slice with negative step",
		expr:        "$[3:0:-2]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["forth","second"]`,
	}
	m["Array slice with negative step on partially overlapping array"] = JsonpathGetCase{
		name:        "Array slice with negative step on partially overlapping array",
		expr:        "$[7:3:-1]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["fifth"]`,
	}
	m["Array slice with negative step and start greater than end"] = JsonpathGetCase{
		name:        "Array slice with negative step and start greater than end",
		expr:        "$[0:3:-2]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `[]`,
	}
	m["Array slice with negative step only"] = JsonpathGetCase{
		name:        "Array slice with negative step only",
		expr:        "$[::-2]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["fifth","third","first"]`,
	}
	m["Array slice with open end"] = JsonpathGetCase{
		name:        "Array slice with open end",
		expr:        "$[1:]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["second","third","forth","fifth"]`,
	}
	m["Array slice with open end and negative step"] = JsonpathGetCase{
		name:        "Array slice with open end and negative step",
		expr:        "$[3::-1]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["forth","third","second","first"]`,
	}
	m["Array slice with open start"] = JsonpathGetCase{
		name:        "Array slice with open start",
		expr:        "$[:2]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["first","second"]`,
	}
	m["Array slice with open start and end"] = JsonpathGetCase{
		name:        "Array slice with open start and end",
		expr:        "$[:]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["first","second","third","forth","fifth"]`,
	}
	m["Array slice with open start and end on object"] = JsonpathGetCase{
		name:        "Array slice with open start and end on object",
		expr:        "$[:]",
		data:        `{":": 42, "more": "string"}`,
		expectation: `[]`,
	}
	m["Array slice with open start and end and step empty"] = JsonpathGetCase{
		name:        "Array slice with open start and end and step empty",
		expr:        "$[::]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["first","second","third","forth","fifth"]`,
	}
	m["Array slice with open start and negative step"] = JsonpathGetCase{
		name:        "Array slice with open start and negative step",
		expr:        "$[:2:-1]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["fifth","forth"]`,
	}
	m["Array slice with positive start and negative end and range of -1"] = JsonpathGetCase{
		name:        "Array slice with positive start and negative end and range of -1",
		expr:        "$[3:-4]",
		data:        `[2, "a", 4, 5, 100, "nice"]`,
		expectation: `[]`,
	}
	m["Array slice with positive start and negative end and range of 0"] = JsonpathGetCase{
		name:        "Array slice with positive start and negative end and range of 0",
		expr:        "$[3:-3]",
		data:        `[2, "a", 4, 5, 100, "nice"]`,
		expectation: `[]`,
	}
	m["Array slice with positive start and negative end and range of 1"] = JsonpathGetCase{
		name:        "Array slice with positive start and negative end and range of 1",
		expr:        "$[3:-2]",
		data:        `[2, "a", 4, 5, 100, "nice"]`,
		expectation: `[5]`,
	}
	m["Array slice with range of -1"] = JsonpathGetCase{
		name:        "Array slice with range of -1",
		expr:        "$[2:1]",
		data:        `["first","second","third","forth"]`,
		expectation: `[]`,
	}
	m["Array slice with range of 0"] = JsonpathGetCase{
		name:        "Array slice with range of 0",
		expr:        "$[0:0]",
		data:        `["first","second","third","forth"]`,
		expectation: `[]`,
	}
	m["Array slice with range of 1"] = JsonpathGetCase{
		name:        "Array slice with range of 1",
		expr:        "$[0:1]",
		data:        `["first","second","third","forth"]`,
		expectation: `["first"]`,
	}
	m["Array slice with start -1 and open end"] = JsonpathGetCase{
		name:        "Array slice with start -1 and open end",
		expr:        "$[-1:]",
		data:        `["first", "second", "third"]`,
		expectation: `["third"]`,
	}
	m["Array slice with start -2 and open end"] = JsonpathGetCase{
		name:        "Array slice with start -2 and open end",
		expr:        "$[-2:]",
		data:        `["first", "second", "third"]`,
		expectation: `["second","third"]`,
	}
	m["Array slice with start large negative number and open end on short array"] = JsonpathGetCase{
		name:        "Array slice with start large negative number and open end on short array",
		expr:        "$[-4:]",
		data:        `["first", "second", "third"]`,
		expectation: `["first","second","third"]`,
	}
	m["Array slice with step"] = JsonpathGetCase{
		name:        "Array slice with step",
		expr:        "$[0:3:2]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["first","third"]`,
	}
	m["Array slice with step 0"] = JsonpathGetCase{
		name:        "Array slice with step 0",
		expr:        "$[0:3:0]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["first","second","third"]`,
	}
	m["Array slice with step 1"] = JsonpathGetCase{
		name:        "Array slice with step 1",
		expr:        "$[0:3:1]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["first","second","third"]`,
	}
	m["Array slice with step and leading zeros"] = JsonpathGetCase{
		name:        "Array slice with step and leading zeros",
		expr:        "$[010:024:010]",
		data:        `[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25]`,
		expectation: `[10,20]`,
	}
	m["Array slice with step but end not aligned"] = JsonpathGetCase{
		name:        "Array slice with step but end not aligned",
		expr:        "$[0:4:2]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["first","third"]`,
	}
	m["Array slice with step empty"] = JsonpathGetCase{
		name:        "Array slice with step empty",
		expr:        "$[1:3:]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["second","third"]`,
	}
	m["Array slice with step only"] = JsonpathGetCase{
		name:        "Array slice with step only",
		expr:        "$[::2]",
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["first","third","fifth"]`,
	}
	m["Bracket notation"] = JsonpathGetCase{
		name:        "Bracket notation",
		expr:        "$['key']",
		data:        `{"key": "value"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation on object without key"] = JsonpathGetCase{
		name:        "Bracket notation on object without key",
		expr:        "$['missing']",
		data:        `{"key": "value"}`,
		expectation: `[]`,
	}
	m["Bracket notation after recursive descent"] = JsonpathGetCase{
		name: "Bracket notation after recursive descent",
		expr: "$..[0]",
		data: `
[
    "first",
    {
        "key": [
            "first nested",
            {
                "more": [
                    {
                        "nested": ["deepest", "second"]
                    },
                    ["more", "values"]
                ]
            }
        ]
    }
]`,
		expectation: `["first","first nested",{"nested":["deepest","second"]},"deepest","more"]`,
	}
	m["Bracket notation with NFC path on NFD key"] = JsonpathGetCase{
		name:        "Bracket notation with NFC path on NFD key",
		expr:        "$['ü']",
		data:        `{"ü": 42}`,
		expectation: `[]`,
	}
	m["Bracket notation with dot"] = JsonpathGetCase{
		name: "Bracket notation with dot",
		expr: "$['two.some']",
		data: `
{
    "one": {"key": "value"},
    "two": {"some": "more", "key": "other value"},
    "two.some": "42"
}`,
		expectation: `["42"]`,
	}
	m["Bracket notation with double quotes"] = JsonpathGetCase{
		name:        "Bracket notation with double quotes",
		expr:        `$["key"]`,
		data:        `{"key": "value"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation with empty path"] = JsonpathGetCase{
		name:        "Bracket notation with empty path",
		expr:        `$[]`,
		data:        `{"": 42, "''": 123, "\"\"": 222}`,
		expectation: `[]`,
	}
	m["Bracket notation with empty string"] = JsonpathGetCase{
		name:        "Bracket notation with empty string",
		expr:        `$['']`,
		data:        `{"": 42, "''": 123, "\"\"": 222}`,
		expectation: `[42]`,
	}
	m["Bracket notation with empty string doubled quoted"] = JsonpathGetCase{
		name:        "Bracket notation with empty string doubled quoted",
		expr:        `$[""]`,
		data:        `{"": 42, "''": 123, "\"\"": 222}`,
		expectation: `[42]`,
	}
	m["Bracket notation with negative number on short array"] = JsonpathGetCase{
		name:        "Bracket notation with negative number on short array",
		expr:        `$[-2]`,
		data:        `["one element"]`,
		expectation: `[]`,
	}
	m["Bracket notation with number"] = JsonpathGetCase{
		name:        "Bracket notation with number",
		expr:        `$[2]`,
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["third"]`,
	}
	m["Bracket notation with number on object"] = JsonpathGetCase{
		name:        "Bracket notation with number on object",
		expr:        `$[0]`,
		data:        `{"0": "value"}`,
		expectation: `[]`,
	}
	m["Bracket notation with number on short array"] = JsonpathGetCase{
		name:        "Bracket notation with number on short array",
		expr:        `$[1]`,
		data:        `["one element"]`,
		expectation: `[]`,
	}
	m["Bracket notation with number on string"] = JsonpathGetCase{
		name:        "Bracket notation with number on string",
		expr:        `$[0]`,
		data:        `"Hello World"`,
		expectation: `[]`,
	}
	m["Bracket notation with number after dot notation with wildcard on nested arrays with different length"] = JsonpathGetCase{
		name:        "Bracket notation with number after dot notation with wildcard on nested arrays with different length",
		expr:        `$.*[1]`,
		data:        `[[1], [2,3]]`,
		expectation: `[3]`,
	}
	m["Bracket notation with number -1"] = JsonpathGetCase{
		name:        "Bracket notation with number -1",
		expr:        `$[-1]`,
		data:        `["first", "second", "third"]`,
		expectation: `["third"]`,
	}
	m["Bracket notation with number -1 on empty array"] = JsonpathGetCase{
		name:        "Bracket notation with number -1 on empty array",
		expr:        `$[-1]`,
		data:        `[]`,
		expectation: `[]`,
	}
	m["Bracket notation with number 0"] = JsonpathGetCase{
		name:        "Bracket notation with number 0",
		expr:        `$[0]`,
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `["first"]`,
	}
	m["Bracket notation with quoted array slice literal"] = JsonpathGetCase{
		name:        "Bracket notation with quoted array slice literal",
		expr:        `$[':']`,
		data:        `{":": "value","another": "entry"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation with quoted closing bracket literal"] = JsonpathGetCase{
		name:        "Bracket notation with quoted closing bracket literal",
		expr:        `$[']']`,
		data:        `{"]": 42}`,
		expectation: `[42]`,
	}
	m["Bracket notation with quoted current object literal"] = JsonpathGetCase{
		name:        "Bracket notation with quoted current object literal",
		expr:        `$['@']`,
		data:        `{"@": "value","another": "entry"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation with quoted dot literal"] = JsonpathGetCase{
		name:        "Bracket notation with quoted dot literal",
		expr:        `$['.']`,
		data:        `{".": "value","another": "entry"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation with quoted dot wildcard"] = JsonpathGetCase{
		name:        "Bracket notation with quoted dot wildcard",
		expr:        `$['.*']`,
		data:        `{"key": 42, ".*": 1, "": 10}`,
		expectation: `[1]`,
	}
	m["Bracket notation with quoted dot wildcard"] = JsonpathGetCase{
		name:        "Bracket notation with quoted dot wildcard",
		expr:        `$['"']`,
		data:        `{"\"": "value","another": "entry"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation with quoted escaped backslash"] = JsonpathGetCase{
		name:        "Bracket notation with quoted escaped backslash",
		expr:        `$['\\']`,
		data:        `{"\\":"value"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation with quoted escaped single quote"] = JsonpathGetCase{
		name:        "Bracket notation with quoted escaped single quote",
		expr:        `$['\'']`,
		data:        `{"'":"value"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation with quoted number on object"] = JsonpathGetCase{
		name:        "Bracket notation with quoted number on object",
		expr:        `$['0']`,
		data:        `{"0": "value"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation with quoted root literal"] = JsonpathGetCase{
		name:        "Bracket notation with quoted root literal",
		expr:        `$['$']`,
		data:        `{"$": "value","another": "entry"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation with quoted special characters combined"] = JsonpathGetCase{
		name:        "Bracket notation with quoted special characters combined",
		expr:        `$[':@."$,*\'\\']`,
		data:        `{":@.\"$,*'\\": 42}`,
		expectation: `[42]`,
	}
	m["Bracket notation with quoted string and unescaped single quote"] = JsonpathGetCase{
		name:        "Bracket notation with quoted string and unescaped single quote",
		expr:        `$['single'quote']`,
		data:        `{"single'quote":"value"}`,
		expectation: `[]`,
		isErrorCase: true,
	}
	m["Bracket notation with quoted union literal"] = JsonpathGetCase{
		name:        "Bracket notation with quoted union literal",
		expr:        `$[',']`,
		data:        `{",": "value","another": "entry"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation with quoted wildcard literal"] = JsonpathGetCase{
		name:        "Bracket notation with quoted wildcard literal",
		expr:        `$['*']`,
		data:        `{"*": "value","another": "entry"}`,
		expectation: `["value"]`,
	}
	m["Bracket notation with quoted wildcard literal on object without key"] = JsonpathGetCase{
		name:        "Bracket notation with quoted wildcard literal on object without key",
		expr:        `$['*']`,
		data:        `{"another": "entry"}`,
		expectation: `[]`,
	}
	m["Bracket notation with spaces"] = JsonpathGetCase{
		name:        "Bracket notation with spaces",
		expr:        `$[ 'a' ]`,
		data:        `{" a": 1, "a": 2, " a ": 3, "a ": 4, " 'a' ": 5, " 'a": 6, "a' ": 7, " \"a\" ": 8, "\"a\"": 9}`,
		expectation: `[2]`,
	}
	m["Bracket notation with string including dot wildcard"] = JsonpathGetCase{
		name:        "Bracket notation with string including dot wildcard",
		expr:        `$['ni.*']`,
		data:        `{"nice": 42, "ni.*": 1, "mice": 100}`,
		expectation: `[1]`,
	}
	m["Bracket notation with two literals separated by dot"] = JsonpathGetCase{
		name: "Bracket notation with two literals separated by dot",
		expr: `$['two'.'some']`,
		data: `
{
    "one": {"key": "value"},
    "two": {"some": "more", "key": "other value"},
    "two.some": "42",
    "two'.'some": "43"
}`,
		expectation: `["43"]`,
	}
	m["Bracket notation with two literals separated by dot without quotes"] = JsonpathGetCase{
		name: "Bracket notation with two literals separated by dot without quotes",
		expr: `$[two.some]`,
		data: `
{
    "one": {"key": "value"},
    "two": {"some": "more", "key": "other value"},
    "two.some": "42"
}`,
		isErrorCase: true,
	}
	m["Bracket notation with wildcard on array"] = JsonpathGetCase{
		name: "Bracket notation with wildcard on array",
		expr: `$[*]`,
		data: `
[
    "string",
    42,
    {
        "key": "value"
    },
    [0, 1]
]`,
		expectation: `["string",42,{"key":"value"},[0,1]]`,
	}
	m["Bracket notation with wildcard on empty array"] = JsonpathGetCase{
		name:        "Bracket notation with wildcard on empty array",
		expr:        `$[*]`,
		data:        `[]`,
		expectation: `[]`,
	}
	m["Bracket notation with wildcard on empty object"] = JsonpathGetCase{
		name:        "Bracket notation with wildcard on empty object",
		expr:        `$[*]`,
		data:        `{}`,
		expectation: `[]`,
	}
	m["Bracket notation with wildcard on null value array"] = JsonpathGetCase{
		name:        "Bracket notation with wildcard on null value array",
		expr:        `$[*]`,
		data:        `[40,null,42]`,
		expectation: `[40,null,42]`,
	}
	m["Bracket notation with wildcard on object"] = JsonpathGetCase{
		name: "Bracket notation with wildcard on object",
		expr: `$[*]`,
		data: `
{
    "some": "string",
    "int": 42,
    "object": {
        "key": "value"
    },
    "array": [0, 1]
}`,
		expectation: `[42,{"key":"value"},[0,1],"string"]`,
	}
	m["Bracket notation with wildcard after array slice"] = JsonpathGetCase{
		name:        "Bracket notation with wildcard after array slice",
		expr:        `$[0:2][*]`,
		data:        `[[1, 2], ["a", "b"], [0, 0]]`,
		expectation: `[1,2,"a","b"]`,
	}
	m["Bracket notation with wildcard after dot notation after bracket notation with wildcard"] = JsonpathGetCase{
		name:        "Bracket notation with wildcard after dot notation after bracket notation with wildcard",
		expr:        `$[*].bar[*]`,
		data:        `[{"bar": [42]}]`,
		expectation: `[42]`,
	}
	m["Bracket notation with wildcard after recursive descent"] = JsonpathGetCase{
		name: "Bracket notation with wildcard after recursive descent",
		expr: `$..[*]`,
		data: `{
    "key": "value",
    "another key": {
        "complex": "string",
        "primitives": [0, 1]
    }
}`,
		expectation: `["string","value",0,1,[0,1],{"complex":"string","primitives":[0,1]}]`,
	}
	m["Bracket notation without quotes"] = JsonpathGetCase{
		name:        "Bracket notation without quotes",
		expr:        `$[key]`,
		data:        `{"key": "value"}`,
		isErrorCase: true,
	}
	m["Current with dot notation"] = JsonpathGetCase{
		name:        "Current with dot notation",
		expr:        `@.a`,
		data:        `{"a": 1}`,
		expectation: `[1]`,
	}
	m["Dot bracket notation"] = JsonpathGetCase{
		name:        "Dot bracket notation",
		expr:        `$.['key']`,
		data:        `{"key": "value","other": {"key": [{"key": 42}]}}`,
		expectation: `[]`,
	}
	m["Dot bracket notation with double quotes"] = JsonpathGetCase{
		name:        "Dot bracket notation with double quotes",
		expr:        `$.["key"]`,
		data:        `{"key": "value","other": {"key": [{"key": 42}]}}`,
		expectation: `[]`,
	}
	m["Dot bracket notation without quotes"] = JsonpathGetCase{
		name:        "Dot bracket notation without quotes",
		expr:        `$.[key]`,
		data:        `{"key": "value","other": {"key": [{"key": 42}]}}`,
		isErrorCase: true,
	}
	m["Dot notation"] = JsonpathGetCase{
		name:        "Dot notation",
		expr:        `$.key`,
		data:        `{"key": "value"}`,
		expectation: `["value"]`,
	}
	m["Dot notation on array"] = JsonpathGetCase{
		name:        "Dot notation on array",
		expr:        `$.key`,
		data:        `[0, 1]`,
		isErrorCase: true,
	}
	m["Dot notation on array value"] = JsonpathGetCase{
		name:        "Dot notation on array value",
		expr:        `$.key`,
		data:        `{"key": ["first", "second"]}`,
		expectation: `[["first","second"]]`,
	}
	m["Dot notation on array with containing object matching key"] = JsonpathGetCase{
		name:        "Dot notation on array with containing object matching key",
		expr:        `$.id`,
		data:        `[{"id": 2}]`,
		isErrorCase: true,
	}
	m["Dot notation on empty object value"] = JsonpathGetCase{
		name:        "Dot notation on empty object value",
		expr:        `$.key`,
		data:        `{"key": {}}`,
		expectation: `[{}]`,
	}
	m["Dot notation on null value"] = JsonpathGetCase{
		name:        "Dot notation on null value",
		expr:        `$.key`,
		data:        `{"key": null}`,
		expectation: `[null]`,
	}
	m["Dot notation on object without key"] = JsonpathGetCase{
		name:        "Dot notation on object without key",
		expr:        `$.missing`,
		data:        `{"key": "value"}`,
		expectation: `[]`,
	}
	m["Dot notation after array slice"] = JsonpathGetCase{
		name:        "Dot notation after array slice",
		expr:        `$[0:2].key`,
		data:        `[{"key": "ey"}, {"key": "bee"}, {"key": "see"}]`,
		expectation: `["ey","bee"]`,
	}
	m["Dot notation after bracket notation after recursive descent"] = JsonpathGetCase{
		name: "Dot notation after bracket notation after recursive descent",
		expr: `$..[1].key`,
		data: `
{
  "k": [{"key": "some value"}, {"key": 42}],
  "kk": [[{"key": 100}, {"key": 200}, {"key": 300}], [{"key": 400}, {"key": 500}, {"key": 600}]],
  "key": [0, 1]
}`,
		expectation: `[200,42,500]`,
	}
	m["Dot notation after bracket notation with wildcard"] = JsonpathGetCase{
		name:        "Dot notation after bracket notation with wildcard",
		expr:        `$[*].a`,
		data:        `[{"a": 1},{"a": 1}]`,
		expectation: `[1,1]`,
	}
	m["Dot notation after bracket notation with wildcard on one matching"] = JsonpathGetCase{
		name:        "Dot notation after bracket notation with wildcard on one matching",
		expr:        `$[*].a`,
		data:        `[{"a": 1}]`,
		expectation: `[1]`,
	}
	m["Dot notation after bracket notation with wildcard on some matching"] = JsonpathGetCase{
		name:        "Dot notation after bracket notation with wildcard on some matching",
		expr:        `$[*].a`,
		data:        `[{"a": 1},{"b": 1}]`,
		expectation: `[1]`,
	}
	m["Dot notation after filter expression"] = JsonpathGetCase{
		name:        "Dot notation after filter expression",
		expr:        `$[?(@.id==42)].name`,
		data:        `[{"id": 42, "name": "forty-two"}, {"id": 1, "name": "one"}]`,
		expectation: `["forty-two"]`,
	}
	m["Dot notation after recursive descent"] = JsonpathGetCase{
		name: "Dot notation after recursive descent",
		expr: `$..key`,
		data: `
{
    "object": {
        "key": "value",
        "array": [
            {"key": "something"},
            {"key": {"key": "russian dolls"}}
        ]
    },
    "key": "top"
}`,
		expectation: `["russian dolls","something","top","value",{"key":"russian dolls"}]`,
	}
	m["Dot notation after recursive descent after dot notation"] = JsonpathGetCase{
		name: "Dot notation after recursive descent after dot notation",
		expr: `$.store..price`,
		data: `
{
  "store": {
    "book": [
      {
        "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      {
        "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99
      },
      {
        "category": "fiction",
        "author": "Herman Melville",
        "title": "Moby Dick",
        "isbn": "0-553-21311-3",
        "price": 8.99
      },
      {
        "category": "fiction",
        "author": "J. R. R. Tolkien",
        "title": "The Lord of the Rings",
        "isbn": "0-395-19395-8",
        "price": 22.99
      }
    ],
    "bicycle": {
      "color": "red",
      "price": 19.95
    }
  }
}`,
		expectation: `[12.99,19.95,22.99,8.95,8.99]`,
	}
	m["Dot notation after recursive descent with extra dot"] = JsonpathGetCase{
		name: "Dot notation after recursive descent with extra dot",
		expr: `$...key`,
		data: `
{
    "object": {
        "key": "value",
        "array": [
            {"key": "something"},
            {"key": {"key": "russian dolls"}}
        ]
    },
    "key": "top"
}`,
		expectation: `["russian dolls","something","top","value",{"key":"russian dolls"}]`,
	}
	m["Dot notation after union"] = JsonpathGetCase{
		name:        "Dot notation after union",
		expr:        `$[0,2].key`,
		data:        `[{"key": "ey"}, {"key": "bee"}, {"key": "see"}]`,
		expectation: `["ey","see"]`,
	}
	m["Dot notation after union with keys"] = JsonpathGetCase{
		name: "Dot notation after union with keys",
		expr: `$['one','three'].key`,
		data: `
{
    "one": {"key": "value"},
    "two": {"k": "v"},
    "three": {"some": "more", "key": "other value"}
}`,
		expectation: `["value","other value"]`,
	}
	m["Dot notation with dash"] = JsonpathGetCase{
		name: "Dot notation with dash",
		expr: `$.key-dash`,
		data: `
{
  "key": 42,
  "key-": 43,
  "-": 44,
  "dash": 45,
  "-dash": 46,
  "": 47,
  "key-dash": "value",
  "something": "else"
}`,
		expectation: `["value"]`,
	}
	m["Dot notation with double quotes"] = JsonpathGetCase{
		name:        "Dot notation with double quotes",
		expr:        `$."key"`,
		data:        `{"key": "value","\"key\"": 42}`,
		expectation: `[42]`,
	}
	m["Dot notation with double quotes after recursive descent"] = JsonpathGetCase{
		name: "Dot notation with double quotes after recursive descent",
		expr: `$.."key"`,
		data: `
{
  "object": {
    "key": "value",
    "\"key\"": 100,
    "array": [
      {"key": "something", "\"key\"": 0},
      {"key": {"key": "russian dolls"}, "\"key\"": {"\"key\"": 99}}
    ]
  },
  "key": "top",
  "\"key\"": 42
}`,
		expectation: `[42,100,0,{"\"key\"":99},99]`,
	}
	m["Dot notation with empty path"] = JsonpathGetCase{
		name:        "Dot notation with empty path",
		expr:        `$.`,
		data:        `{"key": 42, "": 9001, "''": "nice"}`,
		expectation: `[9001]`,
	}
	m["Dot notation with key named in"] = JsonpathGetCase{
		name:        "Dot notation with key named in",
		expr:        `$.in`,
		data:        `{"in": "value"}`,
		expectation: `["value"]`,
	}
	m["Dot notation with key named length"] = JsonpathGetCase{
		name:        "Dot notation with key named length",
		expr:        `$.length`,
		data:        `{"length": "value"}`,
		expectation: `["value"]`,
	}
	m["Dot notation with key named length on array"] = JsonpathGetCase{
		name:        "Dot notation with key named length on array",
		expr:        `$.length`,
		data:        `[4, 5, 6]`,
		expectation: `[]`,
	}
	m["Dot notation with key named null"] = JsonpathGetCase{
		name:        "Dot notation with key named null",
		expr:        `$.null`,
		data:        `{"null": "value"}`,
		expectation: `["value"]`,
	}
	m["Dot notation with key named true"] = JsonpathGetCase{
		name:        "Dot notation with key named true",
		expr:        `$.true`,
		data:        `{"true": "value"}`,
		expectation: `["value"]`,
	}
	m["Dot notation with key root literal"] = JsonpathGetCase{
		name:        "Dot notation with key root literal",
		expr:        `$.$`,
		data:        `{"$": "value"}`,
		expectation: `[]`,
	}
	m["Dot notation with non ASCII key"] = JsonpathGetCase{
		name:        "Dot notation with key root literal",
		expr:        `$.屬性`,
		data:        `{"屬性": "value"}`,
		expectation: `["value"]`,
	}
	m["Dot notation with number"] = JsonpathGetCase{
		name:        "Dot notation with number",
		expr:        `$.2`,
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `[]`,
	}
	m["Dot notation with number on object"] = JsonpathGetCase{
		name:        "Dot notation with number on object",
		expr:        `$.2`,
		data:        `{"a": "first", "2": "second", "b": "third"}`,
		expectation: `["second"]`,
	}
	m["Dot notation with number -1"] = JsonpathGetCase{
		name:        "Dot notation with number -1",
		expr:        `$.-1`,
		data:        `["first", "second", "third", "forth", "fifth"]`,
		expectation: `[]`,
	}
	m["Dot notation with single quotes"] = JsonpathGetCase{
		name:        "Dot notation with single quotes",
		expr:        `$.'key'`,
		data:        `{"key": "value","'key'": 42}`,
		expectation: `[42]`,
	}
	m["Dot notation with single quotes after recursive descent"] = JsonpathGetCase{
		name: "Dot notation with single quotes after recursive descent",
		expr: `$..'key'`,
		data: `
{
  "object": {
    "key": "value",
    "'key'": 100,
    "array": [
      {"key": "something", "'key'": 0},
      {"key": {"key": "russian dolls"}, "'key'": {"'key'": 99}}
    ]
  },
  "key": "top",
  "'key'": 42
}`,
		expectation: `[42,100,0,{"'key'":99},99]`,
	}
	m["Dot notation with single quotes and dot"] = JsonpathGetCase{
		name:        "Dot notation with single quotes and dot",
		expr:        `$.'some.key'`,
		data:        `{"some.key": 42, "some": {"key": "value"}, "'some.key'": 43, "'some": {"key'": 0}}`,
		expectation: `[0]`,
	}
	m["Dot notation with space padded key"] = JsonpathGetCase{
		name:        "Dot notation with space padded key",
		expr:        `$. a`,
		data:        `{" a": 1, "a": 2, " a ": 3, "": 4}`,
		isErrorCase: true,
	}
	m["Dot notation with wildcard on array"] = JsonpathGetCase{
		name: "Dot notation with wildcard on array",
		expr: `$.*`,
		data: `
[
    "string",
    42,
    {
        "key": "value"
    },
    [0, 1]
]`,
		expectation: `["string",42,{"key": "value"},[0,1]]`,
	}
	m["Dot notation with wildcard on empty array"] = JsonpathGetCase{
		name:        "Dot notation with wildcard on empty array",
		expr:        `$.*`,
		data:        `[]`,
		expectation: `[]`,
	}
	m["Dot notation with wildcard on empty object"] = JsonpathGetCase{
		name:        "Dot notation with wildcard on empty object",
		expr:        `$.*`,
		data:        `{}`,
		expectation: `[]`,
	}
	m["Dot notation with wildcard on object"] = JsonpathGetCase{
		name: "Dot notation with wildcard on object",
		expr: `$.*`,
		data: `
{
    "some": "string",
    "int": 42,
    "object": {
        "key": "value"
    },
    "array": [0, 1]
}
`,
		expectation: `["string",42,[0,1],{"key": "value"}]`,
	}
	m["Dot notation with wildcard after dot notation after dot notation with wildcard"] = JsonpathGetCase{
		name:        "Dot notation with wildcard after dot notation after dot notation with wildcard",
		expr:        `$.*.bar.*`,
		data:        `[{"bar": [42]}]`,
		expectation: `[42]`,
	}
	m["Dot notation with wildcard after dot notation with wildcard on nested arrays"] = JsonpathGetCase{
		name:        "Dot notation with wildcard after dot notation with wildcard on nested arrays",
		expr:        `$.*.*`,
		data:        `[[1, 2, 3], [4, 5, 6]]`,
		expectation: `[1,2,3,4,5,6]`,
	}
	m["Dot notation with wildcard after recursive descent"] = JsonpathGetCase{
		name: "Dot notation with wildcard after recursive descent",
		expr: `$..*`,
		data: `
{
    "key": "value",
    "another key": {
        "complex": "string",
        "primitives": [0, 1]
    }
}`,
		expectation: `["string","value",0,1,[0,1],{"complex": "string","primitives": [0,1]}]`,
	}
	m["Dot notation with wildcard after recursive descent on null value array"] = JsonpathGetCase{
		name:        "Dot notation with wildcard after recursive descent on null value array",
		expr:        `$..*`,
		data:        `[40,null,42]`,
		expectation: `[40,42,null]`,
	}
	m["Dot notation with wildcard after recursive descent on scalar"] = JsonpathGetCase{
		name:        "Dot notation with wildcard after recursive descent on scalar",
		expr:        `$..*`,
		data:        `42`,
		expectation: `[]`,
	}
	m["Dot notation without dot"] = JsonpathGetCase{
		name:        "Dot notation without dot",
		expr:        `$a`,
		data:        `{"a": 1, "$a": 2}`,
		isErrorCase: true,
	}
	m["Dot notation without root"] = JsonpathGetCase{
		name:        "Dot notation without root",
		expr:        `.key`,
		data:        `{"key": "value"}`,
		expectation: `["value"]`,
	}
	m["Dot notation without root and dot"] = JsonpathGetCase{
		name:        "Dot notation without root and dot",
		expr:        `key`,
		data:        `{"key": "value"}`,
		isErrorCase: true,
	}
	m["Empty"] = JsonpathGetCase{
		name:        "Empty",
		expr:        ``,
		data:        `{"a": 42, "": 21}`,
		isErrorCase: true,
	}
	m["Filter expression on object"] = JsonpathGetCase{
		name:        "Filter expression on object",
		expr:        `$[?(@.key)]`,
		data:        `{"key": 42, "another": {"key": 1}}`,
		expectation: `[{"key": 1}]`,
	}
	m["Filter expression after dot notation with wildcard after recursive descent"] = JsonpathGetCase{
		name: "Filter expression after dot notation with wildcard after recursive descent",
		expr: `$..*[?(@.id>2)]`,
		data: `
[
    {
        "complext": {
            "one": [
                {"name": "first","id": 1},
                {"name": "next","id": 2},
                {"name": "another","id": 3},
                {"name": "more","id": 4}],
                "more": {"name": "next to last","id": 5}
        }
    },
    {"name": "last","id": 6}
]`,
		expectation: `[{"id": 3,"name": "another"},{"id": 4,"name": "more"},{"id": 5,"name": "next to last"}]`,
	}
	m["Filter expression after recursive descent"] = JsonpathGetCase{
		name:        "Filter expression after recursive descent",
		expr:        `$..[?(@.id==2)]`,
		data:        `{"id": 2, "more": [{"id": 2}, {"more": {"id": 2}}, {"id": {"id": 2}}, [{"id": 2}]]}`,
		expectation: `[{"id":2},{"id":2},{"id":2},{"id":2}]`,
	}
	m["Filter expression with addition"] = JsonpathGetCase{
		name:        "Filter expression with addition",
		expr:        `$[?(@.key+50==100)]`,
		data:        `[{"key": 60}, {"key": 50}, {"key": 10}, {"key": -50}, {"key+50": 100}]`,
		expectation: `[{"key+50":100}]`,
	}
}

func TestGetFunction(t *testing.T) {
	testCases := make(map[string]JsonpathGetCase, 0)
	LoadGetCases(&testCases)
	//testCases = map[string]JsonpathGetCase{"": testCases["Filter expression with boolean and operator"]}
	caseCount, failCount := 0, 0
	for _, c := range testCases {
		caseCount++
		jsonObj := ConvertToJsonObj(c.data)
		j, err := New(c.name, c.expr)
		if err != nil && c.isErrorCase {
			t.Log("[✅PASS] " + c.name)
		} else if err != nil {
			t.Errorf("[⛔️parser error] when create jsonpath(%s)=%s: %v", c.name, c.expr, err)
			return
		} else {
			j.InitData(jsonObj)
			jsonpathResult, err := j.Get()
			if err != nil {
				if c.isErrorCase {
					t.Log("[✅PASS] "+c.name, "expected error: "+err.Error())
					continue
				} else if !c.isErrorCase {
					failCount++
					t.Error("[⛔️jsonpath error]️" + err.Error() + " when " + c.name)
					return
				}
			}
			warnMsg := ""
			if len(j.warnings) > 0 {
				sb := strings.Builder{}
				for i, w := range j.warnings {
					sb.WriteString(fmt.Sprintf("%d. %s; ", i+1, w))
				}
				warnMsg = sb.String()
			}

			resultJsonBytes, _ := json.Marshal(jsonpathResult)
			var result, expectation []interface{}
			json.Unmarshal(resultJsonBytes, &result)
			if !c.isErrorCase {
				err = json.Unmarshal([]byte(c.expectation), &expectation)
				if err != nil {
					failCount++
					t.Error("[⛔️FAIL] cannot unmarshal the expectation json of " + c.name + ": " + err.Error())
					continue
				}
			}
			if Equal(result, expectation) {
				passMsg := fmt.Sprint("[✅PASS] " + c.name)
				if warnMsg != "" {
					passMsg = fmt.Sprint(passMsg + " but have a warning❗️️: " + warnMsg)
				}
				t.Log(passMsg)
			} else {
				failCount++
				t.Error("[⛔️FAIL] " + c.name + " is not correct, the result: " + string(resultJsonBytes) + ", the expectation: " + c.expectation)
			}
		}
	}
	t.Logf("SUMMARY: [TOTAL]=%d [✅PASS]=%d [⛔️FAIL]=%d", caseCount, caseCount-failCount, failCount)
}
