package protobuf

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/NoahOrberg/x/protobuf/ast"
	"github.com/NoahOrberg/x/protobuf/parser"
)

func Parse(target string) (*ast.FileSet, error) {
	targetDir, targetFileName := filepath.Split(target)

	path := os.Getenv("PROTO_PATH") +
		":" + os.Getenv("GOPATH") + "/src/github.com/protocolbuffers/protobuf/src" + // TODO: modify it!!!!
		":" + targetDir

	paths := strings.Split(path, ":")

	paths = append([]string{"sample"}, paths...)
	return parser.ParseFiles([]string{targetFileName}, paths)
}
