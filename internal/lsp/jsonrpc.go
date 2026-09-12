package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Result  any         `json:"result,omitempty"`
	Error   *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// Minimal standard LSP structs with zero external dependencies
type Position struct {
	Line      uint32 `json:"line"`
	Character uint32 `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity,omitempty"`
	Source   string `json:"source,omitempty"`
	Message  string `json:"message"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

type CompletionItem struct {
	Label         string  `json:"label"`
	Kind          int     `json:"kind,omitempty"`
	Detail        *string `json:"detail,omitempty"`
	Documentation any     `json:"documentation,omitempty"`
}

type Hover struct {
	Contents any    `json:"contents"`
	Range    *Range `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Server struct {
	version   string
	documents sync.Map // map[string]string
	writerMu  sync.Mutex
	writer    io.Writer
}

func NewServer(version string) *Server {
	return &Server{
		version: version,
		writer:  os.Stdout,
	}
}

func (s *Server) RunStdio() error {
	reader := bufio.NewReader(os.Stdin)
	for {
		// Read headers
		contentLength := -1
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				break
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 && strings.TrimSpace(strings.ToLower(parts[0])) == "content-length" {
				cl, err := strconv.Atoi(strings.TrimSpace(parts[1]))
				if err == nil {
					contentLength = cl
				}
			}
		}

		if contentLength <= 0 {
			continue
		}

		body := make([]byte, contentLength)
		if _, err := io.ReadFull(reader, body); err != nil {
			return err
		}

		var req Request
		if err := json.Unmarshal(body, &req); err != nil {
			continue
		}

		s.handleMessage(&req)
	}
}

func (s *Server) sendResponse(id *json.RawMessage, result any, respErr *ResponseError) {
	if id == nil {
		return
	}
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
		Error:   respErr,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	s.writePayload(data)
}

func (s *Server) sendNotification(method string, params any) {
	notif := Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(notif)
	if err != nil {
		return
	}
	s.writePayload(data)
}

func (s *Server) writePayload(payload []byte) {
	s.writerMu.Lock()
	defer s.writerMu.Unlock()
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(payload))
	s.writer.Write([]byte(header))
	s.writer.Write(payload)
}
