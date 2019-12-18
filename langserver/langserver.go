package langserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/NoahOrberg/protobuf_langserver/log"
	"github.com/NoahOrberg/protobuf_langserver/protobuf"
	"github.com/NoahOrberg/x/protobuf/ast"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type handler struct {
	mu *sync.Mutex

	// chached AST
	ast *ast.FileSet

	initReq *lsp.InitializeParams
}

func NewHandler() jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError(
		(&handler{
			mu:      new(sync.Mutex),
			ast:     nil,
			initReq: nil,
		}).handle)
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch req.Method {
	case "initialize":
		return h.init(ctx, conn, req)
	case "initialized":
		log.L().Info("invoked initialized method")
		return nil, nil
	case "textDocument/didOpen":
		log.L().Info("invoked textDocument/didOpen method")
		// TODO: Fix, because every time editor open it
		params, err := toDocPosParams(req.Params)
		if err != nil {
			var params string
			if req.Params != nil {
				params = string(*req.Params)
			}
			log.L().Error("failed toDocPosParams", zap.String("params", params))
		}

		// NOTE: if target file is already parsed, ignore it
		if h.ast != nil {
			for _, f := range h.ast.Files {
				if strings.Contains(string(params.TextDocument.URI), f.Name) {
					return nil, nil
				}
			}
		}

		h.ast, err = parse(string(params.TextDocument.URI))
		if err != nil {
			log.L().Error("failed parse uri",
				zap.String("uri", string(params.TextDocument.URI)))
		}
		log.L().Info("success! parsed uri", zap.String("uri", string(params.TextDocument.URI)))

		return nil, nil
	case "textDocument/didChange":
		log.L().Info("invoked textDocument/didChange method")

		// TODO: Fix, because every time editor change it
		params, err := toDocPosParams(req.Params)
		if err != nil {
			var params string
			if req.Params != nil {
				params = string(*req.Params)
			}
			log.L().Error("failed toDocPosParams", zap.String("params", params))
		}
		h.ast, err = parse(string(params.TextDocument.URI))
		if err != nil {
			log.L().Error("failed parse uri",
				zap.String("uri", string(params.TextDocument.URI)))
		}
		log.L().Info("success! parsed uri", zap.String("uri", string(params.TextDocument.URI)))

		return nil, nil
	case "textDocument/definition":
		log.L().Info("invoked textDocument/definition method")

		params, err := toDocPosParams(req.Params)
		if err != nil {
			var params string
			if req.Params != nil {
				params = string(*req.Params)
			}
			log.L().Error("failed toDocPosParams", zap.String("params", params))
			return nil, err
		}

		log.L().Info("toDocPosParams", zap.Any("params", params), zap.Error(err))

		resp, err := h.handleDefinition(ctx, conn, req, params)
		if err != nil {
			log.L().Error("handleDefinition", zap.Any("resp", resp), zap.Error(err))
			return nil, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInternalError,
				Message: fmt.Sprintf("failed resolve definition: %v", err),
			}
		}
		log.L().Info("handleDefinition", zap.Any("resp", resp), zap.Error(err))

		return resp, nil
	}
	return nil, &jsonrpc2.Error{
		Code:    jsonrpc2.CodeMethodNotFound,
		Message: "method is not impl yet",
	}
}

func toDocPosParams(reqParams *json.RawMessage) (*lsp.TextDocumentPositionParams, error) {
	if reqParams == nil {
		return nil, &jsonrpc2.Error{
			Code: jsonrpc2.CodeParseError,
		}
	}
	params := &lsp.TextDocumentPositionParams{}
	if err := json.Unmarshal(*reqParams, params); err != nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInternalError,
			Message: "failed unmarshal json",
		}
	}
	return params, nil
}

func parse(uri string) (*ast.FileSet, error) {
	ast, err := protobuf.Parse(uri)
	if err != nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInternalError,
			Message: fmt.Sprintf("failed parse: %s", uri),
		}
	}
	return ast, nil
}
