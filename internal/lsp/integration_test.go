package lsp_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
	"strings"
	"testing"

	"qwiclang/internal/lsp"
)

func TestLSPInitializeAndHover(t *testing.T) {
	server := lsp.NewServer("v0.1.0-alpha")

	// serverIn reads from clientWriter; serverWriter writes to clientReader
	serverReader, clientWriter := io.Pipe()
	clientReader, serverWriter := io.Pipe()

	server.SetIO(serverReader, serverWriter)

	go func() {
		_ = server.Run()
	}()

	sendRequest := func(method string, id int, params any) {
		bodyMap := map[string]any{
			"jsonrpc": "2.0",
			"id":      id,
			"method":  method,
		}
		if params != nil {
			bodyMap["params"] = params
		}
		data, _ := json.Marshal(bodyMap)
		header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
		clientWriter.Write([]byte(header))
		clientWriter.Write(data)
	}

	sendNotification := func(method string, params any) {
		bodyMap := map[string]any{
			"jsonrpc": "2.0",
			"method":  method,
			"params":  params,
		}
		data, _ := json.Marshal(bodyMap)
		header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
		clientWriter.Write([]byte(header))
		clientWriter.Write(data)
	}

	reader := bufio.NewReader(clientReader)
	readResponse := func() map[string]any {
		tp := textproto.NewReader(reader)
		headers, err := tp.ReadMIMEHeader()
		if err != nil {
			t.Fatalf("failed reading headers: %v", err)
		}
		cl, _ := strconv.Atoi(headers.Get("Content-Length"))
		body := make([]byte, cl)
		io.ReadFull(reader, body)
		var resp map[string]any
		json.Unmarshal(body, &resp)
		return resp
	}

	// 1. Initialize
	sendRequest("initialize", 1, map[string]any{
		"capabilities": map[string]any{},
	})
	initResp := readResponse()
	result := initResp["result"].(map[string]any)
	caps := result["capabilities"].(map[string]any)
	if caps["hoverProvider"] != true {
		t.Fatalf("expected hoverProvider=true, got %v", caps["hoverProvider"])
	}

	// 2. Open document
	docURI := "file:///test.qw"
	sendNotification("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{
			"uri":  docURI,
			"text": "import math\npublic func main() {\n    math.abs(-5.0)\n}",
		},
	})
	// Expect diagnostics notification
	notif := readResponse()
	if notif["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("expected publishDiagnostics notification, got %v", notif["method"])
	}

	// 3. Request hover on math.abs
	sendRequest("textDocument/hover", 2, map[string]any{
		"textDocument": map[string]any{"uri": docURI},
		"position":     map[string]any{"line": 2, "character": 10},
	})
	hoverResp := readResponse()
	hoverRes := hoverResp["result"].(map[string]any)
	contents := hoverRes["contents"].(map[string]any)
	val := contents["value"].(string)
	if !strings.Contains(val, "math.abs") {
		t.Fatalf("expected hover to contain 'math.abs', got: %s", val)
	}

	// 4. Request completion on dot
	sendRequest("textDocument/completion", 3, map[string]any{
		"textDocument": map[string]any{"uri": docURI},
		"position":     map[string]any{"line": 2, "character": 9}, // after "math."
	})
	compResp := readResponse()
	compItems := compResp["result"].([]any)
	if len(compItems) == 0 {
		t.Fatalf("expected completions for math.*, got 0")
	}

	// Shutdown
	sendRequest("shutdown", 4, nil)
	_ = readResponse()
	clientWriter.Close()
	clientReader.Close()
}
