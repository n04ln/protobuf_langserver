package langserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/NoahOrberg/x/protobuf/ast"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type handler struct {
	mu sync.Mutex

	// chached AST
	ast *ast.FileSet

	initReq *lsp.InitializeParams
}

func NewHandler() jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError((&handler{}).handle)
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch req.Method {
	case "initialize":
		return h.init(ctx, conn, req)
	case "initialized":
		return nil, nil
	case "textDocument/didOpen":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		params := &lsp.TextDocumentItem{}
		if err := json.Unmarshal(*req.Params, params); err != nil {
			return nil, err
		}
		return h.didOpen(ctx, params)
	case "textDocument/definition":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		params := &lsp.TextDocumentPositionParams{}
		if err := json.Unmarshal(*req.Params, params); err != nil {
			return nil, err
		}
		return h.handleDefinition(ctx, conn, req, params)
	}
	return nil, fmt.Errorf("method is not impl yet: %s", req.Method)
}
