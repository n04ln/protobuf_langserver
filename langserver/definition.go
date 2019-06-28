package langserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"os"

	"github.com/NoahOrberg/protobuf_langserver/protobuf"
	"github.com/NoahOrberg/x/protobuf/ast"
	"github.com/k0kubun/pp"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *handler) handleDefinition(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request, params *lsp.TextDocumentPositionParams) (interface{}, error) {
	ptrLoc, err := resolve(ctx, params, h.ast)
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
	foundMessage   PosFiler
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
	foundMessage PosFiler
}

func same(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, ss1 := range s1 {
		if ss1 != s2[i] {
			return false
		}
	}
	return true
}

func toPkg(s string) ([]string, string) {
	splittedS := strings.Split(s, ".")
	if len(splittedS) == 0 {
		return []string{}, s
	}
	return splittedS[:len(splittedS)-1], splittedS[len(splittedS)-1]
}

func (v *foundMessageVisitor) Visit(node ast.Node) (w ast.Visitor) {
	// walk for getting specifiedField definition
	switch n := node.(type) {
	case *ast.Message:
		pkg, splitedName := toPkg(v.specifiedField.TypeName)

		log.Println(pkg, node.File().Package)
		log.Println(splitedName, n.Name)

		if splitedName == n.Name &&
			same(pkg, node.File().Package) {
			v.foundMessage = n
		}
		return nil
	case *ast.Enum:
		if v.specifiedField.TypeName == n.Name &&
			same(v.specifiedField.File().Package, node.File().Package) {
			v.foundMessage = n
		}
		return nil
	}

	return v
}

type PosFiler interface {
	Pos() ast.Position
	File() *ast.File
}

func resolve(ctx context.Context, params *lsp.TextDocumentPositionParams, fileSet *ast.FileSet) (*lsp.Location, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Print(err)
		}
	}()

	dir, _ := filepath.Split(string(params.TextDocument.URI))

	v := &specifiedFieldVisitor{
		srcFileName:  string(params.TextDocument.URI),
		srcLine:      params.Position.Line + 1,
		srcCharacter: params.Position.Character + 1,
	}

	for _, file := range fileSet.Files {
		ast.WalkFile(v, file)
	}

	var foundMessage PosFiler
	var nv *foundMessageVisitor // NOTE: avoid error
	if v.foundMessage != nil {
		foundMessage = v.foundMessage
		log.Println(pp.Sprint(foundMessage))
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
		log.Println(2)
		return nil, fmt.Errorf("not found specified field at specifiedFieldVisitor: %s", bs)
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
		return nil, fmt.Errorf("not found specified message at foundMessageVisitor: %s", bs)
	}
	foundMessage = nv.foundMessage

resp:
	log.Println("complete")

	gotLine := foundMessage.Pos().Line - 1      // NOTE: Coz this is 1-based
	gotChar := foundMessage.Pos().Character - 1 // NOTE: Coz this is 1-based

	var fname string
	dirs := append([]string{dir}, protobuf.GetPaths()...)
	for _, d := range dirs {
		delimiter := ""
		if ok := strings.HasSuffix(d, "/"); !ok {
			delimiter = "/"
		}
		fn := d + delimiter + foundMessage.File().Name
		if ok := isExist(fn); ok {
			fname = fn
		}
	}
	if ok := strings.HasPrefix(fname, "file://"); !ok {
		fname = "file://" + fname
	}
	res := lsp.Location{
		URI: lsp.DocumentURI(fname),
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
	log.Println(pp.Sprint(res))
	return &res, nil // fmt.Errorf("%v &&& %v", vb, nvb)
}

func isExist(fname string) bool {
	if f, err := os.Stat(fname); os.IsNotExist(err) || f.IsDir() {
		return false
	} else {
		return true
	}
}
