package types

type Kind int

const (
	Invalid Kind = iota
	Void
	Bool
	Int
	Float
	Nano
	String
	List
	Set
	Dictionary
	Tuple
	Null
	Any
	Struct
	Function
	Range
)

type Type struct {
	Kind       Kind
	Name       string
	Parameters []Type
	ReturnType *Type
}

var (
	InvalidType = Type{Kind: Invalid}
	VoidType    = Type{Kind: Void}
	BoolType    = Type{Kind: Bool}
	IntType     = Type{Kind: Int}
	FloatType   = Type{Kind: Float}
	NanoType    = Type{Kind: Nano}
	StringType  = Type{Kind: String}
	ListType    = Type{Kind: List}
	SetType     = Type{Kind: Set}
	DictType    = Type{Kind: Dictionary}
	TupleType   = Type{Kind: Tuple}
	NullType    = Type{Kind: Null}
	AnyType     = Type{Kind: Any}
	RangeType   = Type{Kind: Range}
)

func StructType(name string) Type {
	return Type{Kind: Struct, Name: name}
}

func FunctionType(parameters []Type, returnType Type) Type {
	result := returnType
	return Type{Kind: Function, Parameters: append([]Type(nil), parameters...), ReturnType: &result}
}

func Lookup(name string) (Type, bool) {
	switch name {
	case "void":
		return VoidType, true
	case "bool":
		return BoolType, true
	case "int":
		return IntType, true
	case "float":
		return FloatType, true
	case "nano":
		return NanoType, true
	case "string":
		return StringType, true
	case "list":
		return ListType, true
	case "set":
		return SetType, true
	case "dictionary", "map":
		return DictType, true
	case "tuple":
		return TupleType, true
	case "any":
		return AnyType, true
	default:
		return InvalidType, false
	}
}

func (typ Type) String() string {
	switch typ.Kind {
	case Void:
		return "void"
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Float:
		return "float"
	case Nano:
		return "nano"
	case String:
		return "string"
	case List:
		return "list"
	case Set:
		return "set"
	case Dictionary:
		return "dictionary"
	case Tuple:
		return "tuple"
	case Null:
		return "null"
	case Any:
		return "any"
	case Struct:
		return typ.Name
	case Function:
		return "func"
	case Range:
		return "range"
	default:
		return "invalid"
	}
}

func (typ Type) IsNumeric() bool {
	return typ.Kind == Int || typ.Kind == Float || typ.Kind == Nano
}

func Compatible(target, value Type) bool {
	if target.Kind == Invalid || value.Kind == Invalid {
		return true
	}
	if target.Kind == Any || value.Kind == Any {
		return true
	}
	if target.Kind == value.Kind {
		if target.Kind == Struct {
			return target.Name == value.Name
		}
		if target.Kind == Function {
			return functionCompatible(target, value)
		}
		return true
	}

	// Allow null to be compatible with any non-void type
	if value.Kind == Null {
		return target.Kind != Void
	}

	// Qwic v0 stores nano as its own type, but decimal literals currently arrive
	// from the lexer as Float. This compatibility allows `const x: nano = 0.1`
	// without pretending full fixed-point literal typing is implemented yet.
	return target.Kind == Nano && value.Kind == Float
}

func functionCompatible(target, value Type) bool {
	if len(target.Parameters) != len(value.Parameters) {
		return false
	}
	for index := range target.Parameters {
		if !Compatible(target.Parameters[index], value.Parameters[index]) || !Compatible(value.Parameters[index], target.Parameters[index]) {
			return false
		}
	}
	if target.ReturnType == nil || value.ReturnType == nil {
		return target.ReturnType == nil && value.ReturnType == nil
	}
	return Compatible(*target.ReturnType, *value.ReturnType)
}
