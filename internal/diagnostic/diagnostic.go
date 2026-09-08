package diagnostic

import (
	"fmt"

	"qwiclang/internal/token"
)

type Severity string

const (
	SeverityError Severity = "error"
)

type Diagnostic struct {
	Severity Severity
	Code     string
	Message  string
	Position token.Position
}

func Error(position token.Position, message string) Diagnostic {
	return Diagnostic{
		Severity: SeverityError,
		Message:  message,
		Position: position,
	}
}

func (diagnostic Diagnostic) Error() string {
	location := fmt.Sprintf("%d:%d", diagnostic.Position.Line, diagnostic.Position.Column)
	if diagnostic.Position.Filename != "" {
		location = fmt.Sprintf("%s:%s", diagnostic.Position.Filename, location)
	}
	if diagnostic.Code != "" {
		return fmt.Sprintf("%s[%s]: %s at %s", diagnostic.Severity, diagnostic.Code, diagnostic.Message, location)
	}
	return fmt.Sprintf("%s: %s at %s", diagnostic.Severity, diagnostic.Message, location)
}
