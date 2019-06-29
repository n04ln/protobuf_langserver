package langserver

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"log"
	"github.com/k0kubun/pp"
)

func (h *handler) init(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	log.Println("invoked initialize method")
	if h.initReq != nil {
		return nil, errors.New("language server is already initialized")
	}
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	var params lsp.InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.initReq = &params

	kind := lsp.TDSKIncremental
	res := lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
				Kind: &kind,
			},
			CompletionProvider:           nil,
			DefinitionProvider:           true,
			TypeDefinitionProvider:       false,
			DocumentFormattingProvider:   false,
			DocumentSymbolProvider:       false,
			HoverProvider:                false,
			ReferencesProvider:           false,
			RenameProvider:               false,
			WorkspaceSymbolProvider:      false,
			ImplementationProvider:       false,
			XWorkspaceReferencesProvider: false,
			XDefinitionProvider:          false,
			XWorkspaceSymbolByProperties: false,
			SignatureHelpProvider:        &lsp.SignatureHelpOptions{TriggerCharacters: []string{"(", ","}},
		},
	}

	log.Println(pp.Sprint(res))

	return res, nil
}
