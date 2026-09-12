package lsp

import (
	"encoding/json"
	"strings"

	qwicformat "qwiclang/internal/format"
	"qwiclang/internal/sema"
)

func (s *Server) handleMessage(req *Request) {
	switch req.Method {
	case "initialize":
		result := map[string]any{
			"capabilities": map[string]any{
				"textDocumentSync": 1, // Full
				"completionProvider": map[string]any{
					"triggerCharacters": []string{"."},
				},
				"hoverProvider":              true,
				"definitionProvider":         true,
				"documentFormattingProvider": true,
			},
			"serverInfo": map[string]any{
				"name":    "qwic-lsp",
				"version": s.version,
			},
		}
		s.sendResponse(req.ID, result, nil)

	case "initialized":
		// No response needed for notifications

	case "shutdown":
		s.sendResponse(req.ID, nil, nil)

	case "exit":
		// client requested exit

	case "textDocument/didOpen":
		var params struct {
			TextDocument struct {
				URI  string `json:"uri"`
				Text string `json:"text"`
			} `json:"textDocument"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			s.documents.Store(params.TextDocument.URI, params.TextDocument.Text)
			s.publishDiagnostics(params.TextDocument.URI, params.TextDocument.Text)
		}

	case "textDocument/didChange":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
			ContentChanges []struct {
				Text string `json:"text"`
			} `json:"contentChanges"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil && len(params.ContentChanges) > 0 {
			text := params.ContentChanges[len(params.ContentChanges)-1].Text
			s.documents.Store(params.TextDocument.URI, text)
			s.publishDiagnostics(params.TextDocument.URI, text)
		}

	case "textDocument/didClose":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			s.documents.Delete(params.TextDocument.URI)
			s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
				URI:         params.TextDocument.URI,
				Diagnostics: []Diagnostic{},
			})
		}

	case "textDocument/formatting":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			edits := s.formatDocument(params.TextDocument.URI)
			s.sendResponse(req.ID, edits, nil)
		} else {
			s.sendResponse(req.ID, []TextEdit{}, nil)
		}

	case "textDocument/completion":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
			Position Position `json:"position"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			items := s.complete(params.TextDocument.URI, params.Position)
			s.sendResponse(req.ID, items, nil)
		} else {
			s.sendResponse(req.ID, []CompletionItem{}, nil)
		}

	case "textDocument/hover":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
			Position Position `json:"position"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			hover := s.hover(params.TextDocument.URI, params.Position)
			s.sendResponse(req.ID, hover, nil)
		} else {
			s.sendResponse(req.ID, nil, nil)
		}

	case "textDocument/definition":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
			Position Position `json:"position"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			loc := s.definition(params.TextDocument.URI, params.Position)
			s.sendResponse(req.ID, loc, nil)
		} else {
			s.sendResponse(req.ID, nil, nil)
		}

	default:
		if req.ID != nil {
			s.sendResponse(req.ID, nil, &ResponseError{
				Code:    -32601,
				Message: "Method not found",
			})
		}
	}
}

func (s *Server) publishDiagnostics(uri string, content string) {
	diagnostics := collectDiagnostics(uri, content)
	s.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

func collectDiagnostics(filename string, content string) []Diagnostic {
	_, checkDiagnostics := sema.Check(filename, content)
	lspDiags := make([]Diagnostic, 0, len(checkDiagnostics))

	for _, diag := range checkDiagnostics {
		line := uint32(0)
		if diag.Position.Line > 0 {
			line = uint32(diag.Position.Line - 1)
		}
		col := uint32(0)
		if diag.Position.Column > 0 {
			col = uint32(diag.Position.Column - 1)
		}

		lspDiags = append(lspDiags, Diagnostic{
			Range: Range{
				Start: Position{Line: line, Character: col},
				End:   Position{Line: line, Character: col + 1},
			},
			Severity: 1, // Error
			Source:   "qwic",
			Message:  diag.Message,
		})
	}

	return lspDiags
}

func (s *Server) formatDocument(uri string) []TextEdit {
	val, ok := s.documents.Load(uri)
	if !ok {
		return []TextEdit{}
	}
	content := val.(string)

	formatted, diags := qwicformat.Source(uri, content)
	if len(diags) > 0 || formatted == content {
		return []TextEdit{}
	}

	lines := strings.Split(content, "\n")
	lastLine := uint32(0)
	lastChar := uint32(0)
	if len(lines) > 0 {
		lastLine = uint32(len(lines) - 1)
		lastChar = uint32(len(lines[len(lines)-1]))
	}

	return []TextEdit{
		{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: lastLine, Character: lastChar},
			},
			NewText: formatted,
		},
	}
}
