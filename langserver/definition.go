package langserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/NoahOrberg/protobuf_langserver/protobuf"
	"github.com/NoahOrberg/x/protobuf/ast"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *handler) handleDefinition(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (interface{}, error) {
	ptrLoc, err := resolve(ctx, params)
	if ptrLoc == nil {
		return nil, err
	}
	return *ptrLoc, err
}

type specifiedFieldVisitor struct {
	// args
	srcFileName           string
	srcLine, srcCharacter int

	// resp
	specifiedField *ast.Field
	foundMessage   *ast.Message
}

func (v *specifiedFieldVisitor) Visit(node ast.Node) (w ast.Visitor) {
	// walk for getting a userspecified field
	fileName, line, character := v.srcFileName, v.srcLine, v.srcCharacter

	switch n := node.(type) {
	case *ast.Method:
		if strings.Contains(fileName, n.File().Name) &&
			n.InTypeNamePosStart.Line == line &&
			n.InTypeNamePosStart.Character <= character && character <= n.InTypeNamePosEnd.Character {
			switch n := n.InType.(type) {
			case *ast.Message:
				v.foundMessage = n
				return nil
			}
		}
		if strings.Contains(fileName, n.File().Name) &&
			n.OutTypeNamePosStart.Line == line &&
			n.OutTypeNamePosStart.Character <= character && character <= n.OutTypeNamePosEnd.Character {
			switch n := n.OutType.(type) {
			case *ast.Message:
				v.foundMessage = n
				return nil
			}
		}
	case *ast.Field:
		log.Println(n.TypeName, " ", n.Name, " ", n.Pos().Line, " ", n.Start.Character, " ", n.End.Character)
		if strings.Contains(fileName, n.File().Name) &&
			n.Pos().Line == line &&
			n.Start.Character <= character && character <= n.End.Character {
			v.specifiedField = n
			return nil
		}
	}
	return v
}

type foundMessageVisitor struct {
	// args
	specifiedField *ast.Field

	// resp
	foundMessage *ast.Message
}

func (v *foundMessageVisitor) Visit(node ast.Node) (w ast.Visitor) {
	// walk for getting specifiedField definition
	switch n := node.(type) {
	case *ast.Message:
		splitedName := strings.Split(v.specifiedField.TypeName, ".")
		if splitedName[len(splitedName)-1] == n.Name { // TODO: Name check only??
			v.foundMessage = n
		}
		return nil
	}

	return v
}

func resolve(ctx context.Context, params lsp.TextDocumentPositionParams) (*lsp.Location, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Print(err)
		}
	}()

	fileSet, err := protobuf.Parse(
		string(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}

	dir, _ := filepath.Split(string(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}

	v := &specifiedFieldVisitor{
		srcFileName:  string(params.TextDocument.URI),
		srcLine:      params.Position.Line + 1,
		srcCharacter: params.Position.Character + 1,
	}

	for _, file := range fileSet.Files {
		ast.WalkFile(v, file)
	}

	var foundMessage *ast.Message
	var nv *foundMessageVisitor // NOTE: avoid error
	if v.foundMessage != nil {
		foundMessage = v.foundMessage
		goto resp
	}

	if v.specifiedField == nil {
		errSt := struct {
			FileName string `json:"filename"`
			Line     int    `json:"line"`
			Char     int    `json:"character"`
		}{
			FileName: v.srcFileName,
			Line:     v.srcLine,
			Char:     v.srcCharacter,
		}
		bs, _ := json.Marshal(errSt)
		return nil, fmt.Errorf("not found specified field: %s", bs)
	}

	nv = &foundMessageVisitor{
		specifiedField: v.specifiedField,
	}
	for _, file := range fileSet.Files {
		ast.WalkFile(nv, file)
	}
	if nv.foundMessage == nil {
		errSt := struct {
			Name      string `json:"name"`
			Line      int    `json:"line"`
			Character int    `json:"character"`
		}{
			Name:      nv.specifiedField.TypeName,
			Line:      nv.specifiedField.Pos().Line,
			Character: nv.specifiedField.Pos().Character,
		}
		bs, _ := json.Marshal(errSt)
		return nil, fmt.Errorf("not found specified message: %s", bs)
	}
	foundMessage = nv.foundMessage

resp:
	gotLine := foundMessage.Position.Line - 1      // NOTE: Coz this is 1-based
	gotChar := foundMessage.Position.Character - 1 // NOTE: Coz this is 1-based
	res := lsp.Location{
		URI: lsp.DocumentURI(dir + foundMessage.File().Name),
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      gotLine,
				Character: gotChar,
			},
			End: lsp.Position{
				Line:      gotLine,
				Character: gotChar,
			},
		},
	}
	// vb, _ := json.Marshal(v)
	// nvb, _ := json.Marshal(nv)
	return &res, nil // fmt.Errorf("%v &&& %v", vb, nvb)
}
