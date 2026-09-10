package token

// Kind identifies the category of lexical token.
type Kind int

const (
	Illegal Kind = iota
	EOF
	Newline

	Identifier
	Integer
	Float
	String
	FString

	Public
	Private
	Const
	Let
	Func
	Return
	If
	Else
	While
	For
	In
	Struct
	Import
	Module
	Turbo
	True
	False
	Null

	Plus
	Minus
	Star
	Slash
	Percent
	Assign
	Equal
	NotEqual
	Less
	LessEqual
	Greater
	GreaterEqual
	And
	Or
	Not

	LParen
	RParen
	LBrace
	RBrace
	LBracket
	RBracket
	Colon
	Comma
	Dot
	Semicolon
)

// Position points to a rune location in a source file. Lines and columns are
// one-based because they are intended for diagnostics.
type Position struct {
	Filename string
	Offset   int
	Line     int
	Column   int
}

// Token is a single lexical item from Qwic source.
type Token struct {
	Kind   Kind
	Lexeme string
	Start  Position
	End    Position
}

var keywords = map[string]Kind{
	"public":  Public,
	"private": Private,
	"const":   Const,
	"let":     Let,
	"func":    Func,
	"return":  Return,
	"if":      If,
	"else":    Else,
	"while":   While,
	"for":     For,
	"in":      In,
	"struct":  Struct,
	"import":  Import,
	"module":  Module,
	"turbo":   Turbo,
	"true":    True,
	"false":   False,
	"null":    Null,
}

// LookupIdentifier returns the keyword kind for text, or Identifier when text
// is not reserved in Phase 1.
func LookupIdentifier(text string) Kind {
	if kind, ok := keywords[text]; ok {
		return kind
	}
	return Identifier
}

func (kind Kind) String() string {
	switch kind {
	case Illegal:
		return "Illegal"
	case EOF:
		return "EOF"
	case Newline:
		return "Newline"
	case Identifier:
		return "Identifier"
	case Integer:
		return "Integer"
	case Float:
		return "Float"
	case String:
		return "String"
	case FString:
		return "FString"
	case Public:
		return "Public"
	case Private:
		return "Private"
	case Const:
		return "Const"
	case Let:
		return "Let"
	case Func:
		return "Func"
	case Return:
		return "Return"
	case If:
		return "If"
	case Else:
		return "Else"
	case While:
		return "While"
	case For:
		return "For"
	case In:
		return "In"
	case Struct:
		return "Struct"
	case Import:
		return "Import"
	case Module:
		return "Module"
	case Turbo:
		return "Turbo"
	case True:
		return "True"
	case False:
		return "False"
	case Null:
		return "Null"
	case Plus:
		return "Plus"
	case Minus:
		return "Minus"
	case Star:
		return "Star"
	case Slash:
		return "Slash"
	case Percent:
		return "Percent"
	case Assign:
		return "Assign"
	case Equal:
		return "Equal"
	case NotEqual:
		return "NotEqual"
	case Less:
		return "Less"
	case LessEqual:
		return "LessEqual"
	case Greater:
		return "Greater"
	case GreaterEqual:
		return "GreaterEqual"
	case And:
		return "And"
	case Or:
		return "Or"
	case Not:
		return "Not"
	case LParen:
		return "LParen"
	case RParen:
		return "RParen"
	case LBrace:
		return "LBrace"
	case RBrace:
		return "RBrace"
	case LBracket:
		return "LBracket"
	case RBracket:
		return "RBracket"
	case Colon:
		return "Colon"
	case Comma:
		return "Comma"
	case Dot:
		return "Dot"
	case Semicolon:
		return "Semicolon"
	default:
		return "Unknown"
	}
}
