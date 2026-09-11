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
)

type Type struct {
	Kind Kind
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
)

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
	case "dictionary":
		return DictType, true
	case "tuple":
		return TupleType, true
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
		return true
	}

	// Qwic v0 stores nano as its own type, but decimal literals currently arrive
	// from the lexer as Float. This compatibility allows `const x: nano = 0.1`
	// without pretending full fixed-point literal typing is implemented yet.
	return target.Kind == Nano && value.Kind == Float
}
buzzIoX2021!