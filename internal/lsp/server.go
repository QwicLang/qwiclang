package lsp

import (
	"strings"
	"sync"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"

	qwicformat "qwiclang/internal/format"
	"qwiclang/internal/sema"
)

const ServerName = "qwic-lsp"

type Server struct {
	version   string
	server    *server.Server
	handler   protocol.Handler
	documents sync.Map // map[protocol.DocumentUri]string
}

func NewServer(version string) *Server {
	s := &Server{
		version: version,
	}

	s.handler = protocol.Handler{
		Initialize:             s.initialize,
		Initialized:            s.initialized,
		Shutdown:               s.shutdown,
		SetTrace:               s.setTrace,
		TextDocumentDidOpen:    s.textDocumentDidOpen,
		TextDocumentDidChange:  s.textDocumentDidChange,
		TextDocumentDidClose:   s.textDocumentDidClose,
		TextDocumentFormatting: s.textDocumentFormatting,
		TextDocumentCompletion: s.textDocumentCompletion,
		TextDocumentHover:      s.textDocumentHover,
		TextDocumentDefinition: s.textDocumentDefinition,
	}

	s.server = server.NewServer(&s.handler, ServerName, false)
	return s
}

func (s *Server) RunStdio() error {
	return s.server.RunStdio()
}

func (s *Server) initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	syncKind := protocol.TextDocumentSyncKindFull
	capabilities := protocol.ServerCapabilities{
		TextDocumentSync: syncKind,
		CompletionProvider: &protocol.CompletionOptions{
			TriggerCharacters: []string{"."},
		},
		HoverProvider:              true,
		DefinitionProvider:         true,
		DocumentFormattingProvider: true,
	}

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    ServerName,
			Version: &s.version,
		},
	}, nil
}

func (s *Server) initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func (s *Server) shutdown(context *glsp.Context) error {
	return nil
}

func (s *Server) setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	return nil
}

func (s *Server) textDocumentDidOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	s.documents.Store(params.TextDocument.URI, params.TextDocument.Text)
	s.publishDiagnostics(context, params.TextDocument.URI, params.TextDocument.Text)
	return nil
}

func (s *Server) textDocumentDidChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) > 0 {
		text := params.ContentChanges[len(params.ContentChanges)-1].(protocol.TextDocumentContentChangeEventWhole).Text
		s.documents.Store(params.TextDocument.URI, text)
		s.publishDiagnostics(context, params.TextDocument.URI, text)
	}
	return nil
}

func (s *Server) textDocumentDidClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	s.documents.Delete(params.TextDocument.URI)
	go context.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: []protocol.Diagnostic{},
	})
	return nil
}

func (s *Server) publishDiagnostics(context *glsp.Context, uri protocol.DocumentUri, content string) {
	go func() {
		diagnostics := collectDiagnostics(uri, content)
		context.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diagnostics,
		})
	}()
}

func collectDiagnostics(filename string, content string) []protocol.Diagnostic {
	_, checkDiagnostics := sema.Check(filename, content)
	lspDiags := make([]protocol.Diagnostic, 0, len(checkDiagnostics))

	for _, diag := range checkDiagnostics {
		severity := protocol.DiagnosticSeverityError

		line := uint32(0)
		if diag.Position.Line > 0 {
			line = uint32(diag.Position.Line - 1)
		}
		col := uint32(0)
		if diag.Position.Column > 0 {
			col = uint32(diag.Position.Column - 1)
		}

		source := "qwic"
		lspDiags = append(lspDiags, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: line, Character: col},
				End:   protocol.Position{Line: line, Character: col + 1},
			},
			Severity: &severity,
			Source:   &source,
			Message:  diag.Message,
		})
	}

	return lspDiags
}

func (s *Server) textDocumentFormatting(context *glsp.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	val, ok := s.documents.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	content := val.(string)

	formatted, diags := qwicformat.Source(params.TextDocument.URI, content)
	if len(diags) > 0 || formatted == content {
		return nil, nil
	}

	lines := strings.Split(content, "\n")
	lastLine := uint32(0)
	lastChar := uint32(0)
	if len(lines) > 0 {
		lastLine = uint32(len(lines) - 1)
		lastChar = uint32(len(lines[len(lines)-1]))
	}

	return []protocol.TextEdit{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: lastLine, Character: lastChar},
			},
			NewText: formatted,
		},
	}, nil
}
