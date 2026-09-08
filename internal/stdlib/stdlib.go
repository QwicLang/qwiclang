package stdlib

import "qwiclang/internal/types"

type Parameter struct {
	Name string
	Type types.Type
}

type Function struct {
	Package     string
	Name        string
	RuntimeName string
	Parameters  []Parameter
	ReturnType  types.Type
}

var functions = []Function{
	{
		Package:     "strings",
		Name:        "length",
		RuntimeName: "qwic_strings_length",
		Parameters:  []Parameter{{Name: "value", Type: types.StringType}},
		ReturnType:  types.IntType,
	},
	{
		Package:     "strings",
		Name:        "empty",
		RuntimeName: "qwic_strings_empty",
		Parameters:  []Parameter{{Name: "value", Type: types.StringType}},
		ReturnType:  types.BoolType,
	},
	{
		Package:     "strings",
		Name:        "trim",
		RuntimeName: "qwic_strings_trim",
		Parameters:  []Parameter{{Name: "value", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "strings",
		Name:        "trimLeft",
		RuntimeName: "qwic_strings_trim_left",
		Parameters:  []Parameter{{Name: "value", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "strings",
		Name:        "trimRight",
		RuntimeName: "qwic_strings_trim_right",
		Parameters:  []Parameter{{Name: "value", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "strings",
		Name:        "upper",
		RuntimeName: "qwic_strings_upper",
		Parameters:  []Parameter{{Name: "value", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "strings",
		Name:        "lower",
		RuntimeName: "qwic_strings_lower",
		Parameters:  []Parameter{{Name: "value", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "strings",
		Name:        "contains",
		RuntimeName: "qwic_strings_contains",
		Parameters: []Parameter{
			{Name: "value", Type: types.StringType},
			{Name: "needle", Type: types.StringType},
		},
		ReturnType: types.BoolType,
	},
	{
		Package:     "strings",
		Name:        "startsWith",
		RuntimeName: "qwic_strings_starts_with",
		Parameters: []Parameter{
			{Name: "value", Type: types.StringType},
			{Name: "prefix", Type: types.StringType},
		},
		ReturnType: types.BoolType,
	},
	{
		Package:     "strings",
		Name:        "endsWith",
		RuntimeName: "qwic_strings_ends_with",
		Parameters: []Parameter{
			{Name: "value", Type: types.StringType},
			{Name: "suffix", Type: types.StringType},
		},
		ReturnType: types.BoolType,
	},
	{
		Package:     "strings",
		Name:        "indexOf",
		RuntimeName: "qwic_strings_index_of",
		Parameters: []Parameter{
			{Name: "value", Type: types.StringType},
			{Name: "needle", Type: types.StringType},
		},
		ReturnType: types.IntType,
	},
}

var functionsByQualifiedName = buildFunctionIndex()

func HasPackage(name string) bool {
	for _, function := range functions {
		if function.Package == name {
			return true
		}
	}
	return false
}

func Functions() []Function {
	return append([]Function(nil), functions...)
}

func LookupFunction(qualifiedName string) (Function, bool) {
	function, ok := functionsByQualifiedName[qualifiedName]
	return function, ok
}

func buildFunctionIndex() map[string]Function {
	index := map[string]Function{}
	for _, function := range functions {
		index[function.Package+"."+function.Name] = function
	}
	return index
}
