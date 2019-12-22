package langserver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/n04ln/protobuf_langserver/log"
	"github.com/n04ln/protobuf_langserver/protobuf"
	"github.com/n04ln/x/protobuf/ast"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

func (h *handler) handleDefinition(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request, params *lsp.TextDocumentPositionParams) (interface{}, error) {
	ptrLoc, err := resolve(ctx, params, h.ast)
	if ptrLoc == nil {
		return nil, err
	}
	return *ptrLoc, err
}

type PosFiler interface {
	Pos() ast.Position
	File() *ast.File
}

type PosFileTypeNamer interface {
	PosFiler
	TypeName() string
}

type InOutType struct {
	pos      ast.Position
	file     *ast.File
	typeName string
}

func (t *InOutType) Pos() ast.Position {
	return t.pos
}
func (t *InOutType) File() *ast.File {
	return t.file
}

func (t *InOutType) TypeName() string {
	return t.typeName
}

type FieldWrap struct {
	field *ast.Field
}

func (f *FieldWrap) Pos() ast.Position {
	return f.field.Pos()
}
func (f *FieldWrap) File() *ast.File {
	return f.File()
}

func (f *FieldWrap) TypeName() string {
	return f.field.TypeName
}

type specifiedFieldVisitor struct {
	// args
	srcFileName           string
	srcLine, srcCharacter int

	// resp
	specifiedField PosFileTypeNamer
	foundMessage   PosFiler
}

func (v *specifiedFieldVisitor) Visit(node ast.Node) (w ast.Visitor) {
	// walk for getting a userspecified field
	fileName, line, character := v.srcFileName, v.srcLine, v.srcCharacter
	switch n := node.(type) {
	case *ast.Method:
		log.L().Info(n.InTypeName + " - " + n.OutTypeName)

		if strings.Contains(fileName, n.File().Name) &&
			n.InTypeNamePosStart.Line == line &&
			n.InTypeNamePosStart.Character <= character && character <= n.InTypeNamePosEnd.Character {

			log.L().Info("Found!")
			v.specifiedField = &InOutType{
				pos:      n.InTypeNamePosStart,
				file:     n.File(),
				typeName: n.InTypeName,
			}
			return nil
		}

		if strings.Contains(fileName, n.File().Name) &&
			n.OutTypeNamePosStart.Line == line &&
			n.OutTypeNamePosStart.Character <= character && character <= n.OutTypeNamePosEnd.Character {

			log.L().Info("Found!")
			v.specifiedField = &InOutType{
				pos:      n.OutTypeNamePosStart,
				file:     n.File(),
				typeName: n.OutTypeName,
			}
			return nil
		}

		// TODO: Fix it.
		// n.InType(type) is *ast.File, therefore NoahOrberg/x/protobuf repository is broken.

		// if strings.Contains(fileName, n.File().Name) &&
		// 	n.InTypeNamePosStart.Line == line &&
		// 	n.InTypeNamePosStart.Character <= character && character <= n.InTypeNamePosEnd.Character {
		// 	switch nn := n.InType.(type) {
		// 	case *ast.Message:
		// 		log.L().Info("Found!")
		// 		v.foundMessage = nn
		// 		return nil
		// 	}
		// }
		//
		// log.L().Info("params", zap.String("arg", fileName), zap.String("walked", n.File().Name))
		//
		// if strings.Contains(fileName, n.File().Name) &&
		// 	n.OutTypeNamePosStart.Line == line &&
		// 	n.OutTypeNamePosStart.Character <= character && character <= n.OutTypeNamePosEnd.Character {
		// 	switch nn := n.OutType.(type) {
		// 	case *ast.Message:
		// 		log.L().Info("Found!")
		// 		v.foundMessage = nn
		// 		return nil
		// 	case *ast.File:
		// 		log.L().Info("Found!")
		// 		stdlog.Println("name is ", nn.Name)
		// 		stdlog.Println("messages is ", nn.Messages)
		// 	}
		// }
	case *ast.Field:
		if strings.Contains(fileName, n.File().Name) &&
			n.Pos().Line == line &&
			n.Start.Character <= character && character <= n.End.Character {

			if nn, ok := n.Type.(*ast.Message); ok {
				// NOTE: for inner message
				v.foundMessage = nn
			} else {
				v.specifiedField = &FieldWrap{
					field: n,
				}
			}
			return nil
		}
	}
	return v
}

type foundMessageVisitor struct {
	// args
	specifiedField PosFileTypeNamer

	// resp
	foundMessage PosFiler
}

func (v *foundMessageVisitor) Visit(node ast.Node) (w ast.Visitor) {
	// walk for getting specifiedField definition
	switch n := node.(type) {
	case *ast.Message:
		pkg, splitedName := toPkg(v.specifiedField.TypeName())

		log.L().Info("print info",
			zap.Strings("pkg", pkg), zap.Strings("nodePackage", node.File().Package))
		log.L().Info("print info",
			zap.String("splitedName", splitedName), zap.String("nName", n.Name))

		if splitedName == n.Name {
			if len(pkg) == 0 { // TODO: why pkg is zero-array? (when same package)
				v.foundMessage = n
			} else if same(pkg, node.File().Package) {
				v.foundMessage = n
			}
		}
		return nil
	case *ast.Enum:
		if v.specifiedField.TypeName() == n.Name &&
			same(v.specifiedField.File().Package, node.File().Package) {
			v.foundMessage = n
		}
		return nil
	}

	return v
}

func resolve(ctx context.Context, params *lsp.TextDocumentPositionParams, fileSet *ast.FileSet) (*lsp.Location, error) {
	defer func() {
		if err := recover(); err != nil {
			log.L().Error("panic occured", zap.Error(err.(error)))
		}
	}()

	dir, _ := filepath.Split(string(params.TextDocument.URI))

	v := &specifiedFieldVisitor{
		srcFileName:  string(params.TextDocument.URI),
		srcLine:      params.Position.Line + 1,
		srcCharacter: params.Position.Character + 1,
	}

	for _, file := range fileSet.Files {
		log.L().Info("start walk", zap.String("fname", file.Name))
		ast.WalkFile(v, file)
		log.L().Info("end walk")
	}

	var foundMessage PosFiler
	var nv *foundMessageVisitor // NOTE: avoid error
	if v.foundMessage != nil {
		foundMessage = v.foundMessage
		//TODO
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
		return nil, fmt.Errorf("not found specified field at specifiedFieldVisitor: %s", bs)
	}

	nv = &foundMessageVisitor{
		specifiedField: v.specifiedField,
	}
	for _, file := range fileSet.Files {
		log.L().Info("start walk", zap.String("fname", file.Name))
		ast.WalkFile(nv, file)
		log.L().Info("end walk")
	}
	if nv.foundMessage == nil {
		errSt := struct {
			Name      string `json:"name"`
			Line      int    `json:"line"`
			Character int    `json:"character"`
		}{
			Name:      nv.specifiedField.TypeName(),
			Line:      nv.specifiedField.Pos().Line,
			Character: nv.specifiedField.Pos().Character,
		}
		bs, _ := json.Marshal(errSt)
		return nil, fmt.Errorf("not found specified message at foundMessageVisitor: %s", bs)
	}
	foundMessage = nv.foundMessage

resp:
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

	// NOTE: "file://" prefix is needed for Language Server Protocol
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
	return &res, nil // fmt.Errorf("%v &&& %v", vb, nvb)
}

func isExist(fname string) bool {
	f, err := os.Stat(fname)
	return !(os.IsNotExist(err) || f.IsDir())
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
