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
	{
		Package:     "time",
		Name:        "now",
		RuntimeName: "qwic_time_now",
		ReturnType:  types.IntType,
	},
	{
		Package:     "time",
		Name:        "sleep",
		RuntimeName: "qwic_time_sleep",
		Parameters:  []Parameter{{Name: "ms", Type: types.IntType}},
		ReturnType:  types.VoidType,
	},
	{
		Package:     "time",
		Name:        "duration",
		RuntimeName: "qwic_time_duration",
		Parameters:  []Parameter{{Name: "start", Type: types.IntType}, {Name: "end", Type: types.IntType}},
		ReturnType:  types.IntType,
	},
	{
		Package:     "fs",
		Name:        "read_file",
		RuntimeName: "qwic_fs_read_file",
		Parameters:  []Parameter{{Name: "path", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "fs",
		Name:        "write_file",
		RuntimeName: "qwic_fs_write_file",
		Parameters:  []Parameter{{Name: "path", Type: types.StringType}, {Name: "content", Type: types.StringType}},
		ReturnType:  types.BoolType,
	},
	{
		Package:     "fs",
		Name:        "exists",
		RuntimeName: "qwic_fs_exists",
		Parameters:  []Parameter{{Name: "path", Type: types.StringType}},
		ReturnType:  types.BoolType,
	},
	{
		Package:     "sync",
		Name:        "mutex_new",
		RuntimeName: "qwic_sync_mutex_new",
		ReturnType:  types.AnyType,
	},
	{
		Package:     "sync",
		Name:        "mutex_lock",
		RuntimeName: "qwic_sync_mutex_lock",
		Parameters:  []Parameter{{Name: "mutex", Type: types.AnyType}},
		ReturnType:  types.VoidType,
	},
	{
		Package:     "sync",
		Name:        "mutex_unlock",
		RuntimeName: "qwic_sync_mutex_unlock",
		Parameters:  []Parameter{{Name: "mutex", Type: types.AnyType}},
		ReturnType:  types.VoidType,
	},
	{
		Package:     "sync",
		Name:        "mutex_free",
		RuntimeName: "qwic_sync_mutex_free",
		Parameters:  []Parameter{{Name: "mutex", Type: types.AnyType}},
		ReturnType:  types.VoidType,
	},
	{
		Package:     "lists",
		Name:        "new",
		RuntimeName: "qwic_lists_new",
		ReturnType:  types.ListType,
	},
	{
		Package:     "lists",
		Name:        "push",
		RuntimeName: "qwic_lists_push",
		Parameters:  []Parameter{{Name: "list", Type: types.ListType}, {Name: "value", Type: types.StringType}},
		ReturnType:  types.VoidType,
	},
	{
		Package:     "lists",
		Name:        "get",
		RuntimeName: "qwic_lists_get",
		Parameters:  []Parameter{{Name: "list", Type: types.ListType}, {Name: "index", Type: types.IntType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "lists",
		Name:        "length",
		RuntimeName: "qwic_lists_length",
		Parameters:  []Parameter{{Name: "list", Type: types.ListType}},
		ReturnType:  types.IntType,
	},
	{
		Package:     "lists",
		Name:        "contains",
		RuntimeName: "qwic_lists_contains",
		Parameters:  []Parameter{{Name: "list", Type: types.ListType}, {Name: "value", Type: types.StringType}},
		ReturnType:  types.BoolType,
	},
	{
		Package:     "lists",
		Name:        "slice",
		RuntimeName: "qwic_lists_slice",
		Parameters:  []Parameter{{Name: "list", Type: types.ListType}, {Name: "start", Type: types.IntType}, {Name: "end", Type: types.IntType}},
		ReturnType:  types.ListType,
	},
	{
		Package:     "sets",
		Name:        "new",
		RuntimeName: "qwic_sets_new",
		ReturnType:  types.SetType,
	},
	{
		Package:     "sets",
		Name:        "add",
		RuntimeName: "qwic_sets_add",
		Parameters:  []Parameter{{Name: "set", Type: types.SetType}, {Name: "value", Type: types.StringType}},
		ReturnType:  types.VoidType,
	},
	{
		Package:     "sets",
		Name:        "contains",
		RuntimeName: "qwic_sets_contains",
		Parameters:  []Parameter{{Name: "set", Type: types.SetType}, {Name: "value", Type: types.StringType}},
		ReturnType:  types.BoolType,
	},
	{
		Package:     "sets",
		Name:        "length",
		RuntimeName: "qwic_sets_length",
		Parameters:  []Parameter{{Name: "set", Type: types.SetType}},
		ReturnType:  types.IntType,
	},
	{
		Package:     "dictionaries",
		Name:        "new",
		RuntimeName: "qwic_dictionaries_new",
		ReturnType:  types.DictType,
	},
	{
		Package:     "dictionaries",
		Name:        "set",
		RuntimeName: "qwic_dictionaries_set",
		Parameters:  []Parameter{{Name: "dictionary", Type: types.DictType}, {Name: "key", Type: types.StringType}, {Name: "value", Type: types.StringType}},
		ReturnType:  types.VoidType,
	},
	{
		Package:     "dictionaries",
		Name:        "get",
		RuntimeName: "qwic_dictionaries_get",
		Parameters:  []Parameter{{Name: "dictionary", Type: types.DictType}, {Name: "key", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "dictionaries",
		Name:        "contains",
		RuntimeName: "qwic_dictionaries_contains",
		Parameters:  []Parameter{{Name: "dictionary", Type: types.DictType}, {Name: "key", Type: types.StringType}},
		ReturnType:  types.BoolType,
	},
	{
		Package:     "dictionaries",
		Name:        "length",
		RuntimeName: "qwic_dictionaries_length",
		Parameters:  []Parameter{{Name: "dictionary", Type: types.DictType}},
		ReturnType:  types.IntType,
	},
	{
		Package:     "tuples",
		Name:        "new2",
		RuntimeName: "qwic_tuples_new2",
		Parameters:  []Parameter{{Name: "first", Type: types.StringType}, {Name: "second", Type: types.StringType}},
		ReturnType:  types.TupleType,
	},
	{
		Package:     "tuples",
		Name:        "first",
		RuntimeName: "qwic_tuples_first",
		Parameters:  []Parameter{{Name: "tuple", Type: types.TupleType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "tuples",
		Name:        "second",
		RuntimeName: "qwic_tuples_second",
		Parameters:  []Parameter{{Name: "tuple", Type: types.TupleType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "tuples",
		Name:        "length",
		RuntimeName: "qwic_tuples_length",
		Parameters:  []Parameter{{Name: "tuple", Type: types.TupleType}},
		ReturnType:  types.IntType,
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
