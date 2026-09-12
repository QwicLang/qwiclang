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
		Package:     "math",
		Name:        "abs",
		RuntimeName: "qwic_math_abs",
		Parameters:  []Parameter{{Name: "x", Type: types.FloatType}},
		ReturnType:  types.FloatType,
	},
	{
		Package:     "math",
		Name:        "min",
		RuntimeName: "qwic_math_min",
		Parameters:  []Parameter{{Name: "a", Type: types.FloatType}, {Name: "b", Type: types.FloatType}},
		ReturnType:  types.FloatType,
	},
	{
		Package:     "math",
		Name:        "max",
		RuntimeName: "qwic_math_max",
		Parameters:  []Parameter{{Name: "a", Type: types.FloatType}, {Name: "b", Type: types.FloatType}},
		ReturnType:  types.FloatType,
	},
	{
		Package:     "math",
		Name:        "sqrt",
		RuntimeName: "qwic_math_sqrt",
		Parameters:  []Parameter{{Name: "x", Type: types.FloatType}},
		ReturnType:  types.FloatType,
	},
	{
		Package:     "math",
		Name:        "pow",
		RuntimeName: "qwic_math_pow",
		Parameters:  []Parameter{{Name: "base", Type: types.FloatType}, {Name: "exp", Type: types.FloatType}},
		ReturnType:  types.FloatType,
	},
	{
		Package:     "math",
		Name:        "floor",
		RuntimeName: "qwic_math_floor",
		Parameters:  []Parameter{{Name: "x", Type: types.FloatType}},
		ReturnType:  types.FloatType,
	},
	{
		Package:     "math",
		Name:        "ceil",
		RuntimeName: "qwic_math_ceil",
		Parameters:  []Parameter{{Name: "x", Type: types.FloatType}},
		ReturnType:  types.FloatType,
	},
	{
		Package:     "math",
		Name:        "round",
		RuntimeName: "qwic_math_round",
		Parameters:  []Parameter{{Name: "x", Type: types.FloatType}},
		ReturnType:  types.FloatType,
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
		Package:     "crypto",
		Name:        "sha256",
		RuntimeName: "qwic_crypto_sha256",
		Parameters:  []Parameter{{Name: "value", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "crypto",
		Name:        "sha512",
		RuntimeName: "qwic_crypto_sha512",
		Parameters:  []Parameter{{Name: "value", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "crypto",
		Name:        "random_bytes",
		RuntimeName: "qwic_crypto_random_bytes",
		Parameters:  []Parameter{{Name: "len", Type: types.IntType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "crypto",
		Name:        "random_int",
		RuntimeName: "qwic_crypto_random_int",
		ReturnType:  types.IntType,
	},
	{
		Package:     "crypto",
		Name:        "hex_encode",
		RuntimeName: "qwic_crypto_hex_encode",
		Parameters:  []Parameter{
			{Name: "bytes", Type: types.StringType},
			{Name: "len", Type: types.IntType},
		},
		ReturnType:  types.StringType,
	},
	{
		Package:     "crypto",
		Name:        "base64_encode",
		RuntimeName: "qwic_crypto_base64_encode",
		Parameters:  []Parameter{
			{Name: "bytes", Type: types.StringType},
			{Name: "len", Type: types.IntType},
		},
		ReturnType:  types.StringType,
	},
	{
		Package:     "net",
		Name:        "tcp_connect",
		RuntimeName: "qwic_net_tcp_connect",
		Parameters:  []Parameter{
			{Name: "host", Type: types.StringType},
			{Name: "port", Type: types.IntType},
		},
		ReturnType:  types.AnyType,
	},
	{
		Package:     "net",
		Name:        "tcp_listen",
		RuntimeName: "qwic_net_tcp_listen",
		Parameters:  []Parameter{{Name: "port", Type: types.StringType}},
		ReturnType:  types.AnyType,
	},
	{
		Package:     "net",
		Name:        "tcp_accept",
		RuntimeName: "qwic_net_tcp_accept",
		Parameters:  []Parameter{{Name: "conn", Type: types.AnyType}},
		ReturnType:  types.AnyType,
	},
	{
		Package:     "net",
		Name:        "tcp_read",
		RuntimeName: "qwic_net_tcp_read",
		Parameters:  []Parameter{
			{Name: "conn", Type: types.AnyType},
			{Name: "max_len", Type: types.IntType},
		},
		ReturnType:  types.StringType,
	},
	{
		Package:     "net",
		Name:        "tcp_write",
		RuntimeName: "qwic_net_tcp_write",
		Parameters:  []Parameter{
			{Name: "conn", Type: types.AnyType},
			{Name: "data", Type: types.StringType},
		},
		ReturnType:  types.IntType,
	},
	{
		Package:     "net",
		Name:        "tcp_close",
		RuntimeName: "qwic_net_tcp_close",
		Parameters:  []Parameter{{Name: "conn", Type: types.AnyType}},
		ReturnType:  types.VoidType,
	},
	{
		Package:     "net",
		Name:        "dns_lookup",
		RuntimeName: "qwic_net_dns_lookup",
		Parameters:  []Parameter{{Name: "host", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "json",
		Name:        "stringify",
		RuntimeName: "qwic_json_stringify",
		Parameters:  []Parameter{{Name: "value", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "json",
		Name:        "parse",
		RuntimeName: "qwic_json_parse",
		Parameters:  []Parameter{{Name: "json", Type: types.StringType}},
		ReturnType:  types.AnyType,
	},
	{
		Package:     "json",
		Name:        "encode",
		RuntimeName: "qwic_json_encode",
		Parameters:  []Parameter{{Name: "value", Type: types.StringType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "json",
		Name:        "decode",
		RuntimeName: "qwic_json_decode",
		Parameters:  []Parameter{{Name: "json", Type: types.StringType}},
		ReturnType:  types.AnyType,
	},
	{
		Package:     "http",
		Name:        "request",
		RuntimeName: "qwic_http_request_new",
		Parameters:  []Parameter{{Name: "url", Type: types.StringType}},
		ReturnType:  types.AnyType,
	},
	{
		Package:     "http",
		Name:        "set_method",
		RuntimeName: "qwic_http_request_set_method",
		Parameters:  []Parameter{
			{Name: "req", Type: types.AnyType},
			{Name: "method", Type: types.StringType},
		},
		ReturnType:  types.VoidType,
	},
	{
		Package:     "http",
		Name:        "set_body",
		RuntimeName: "qwic_http_request_set_body",
		Parameters:  []Parameter{
			{Name: "req", Type: types.AnyType},
			{Name: "body", Type: types.StringType},
		},
		ReturnType:  types.VoidType,
	},
	{
		Package:     "http",
		Name:        "send",
		RuntimeName: "qwic_http_send",
		Parameters:  []Parameter{{Name: "req", Type: types.AnyType}},
		ReturnType:  types.AnyType,
	},
	{
		Package:     "http",
		Name:        "free_request",
		RuntimeName: "qwic_http_request_free",
		Parameters:  []Parameter{{Name: "req", Type: types.AnyType}},
		ReturnType:  types.VoidType,
	},
	{
		Package:     "http",
		Name:        "get_status",
		RuntimeName: "qwic_http_get_status",
		Parameters:  []Parameter{{Name: "response", Type: types.AnyType}},
		ReturnType:  types.IntType,
	},
	{
		Package:     "http",
		Name:        "get_body",
		RuntimeName: "qwic_http_get_body",
		Parameters:  []Parameter{{Name: "response", Type: types.AnyType}},
		ReturnType:  types.StringType,
	},
	{
		Package:     "http",
		Name:        "free_response",
		RuntimeName: "qwic_http_free_response",
		Parameters:  []Parameter{{Name: "response", Type: types.AnyType}},
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
