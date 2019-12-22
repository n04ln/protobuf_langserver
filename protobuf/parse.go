package protobuf

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/n04ln/x/protobuf/ast"
	"github.com/n04ln/x/protobuf/parser"
)

var (
	targetDir, targetFileName string
)

func Parse(target string) (*ast.FileSet, error) {
	targetDir, targetFileName = filepath.Split(target)

	targetDir = strings.TrimPrefix(targetDir, "file://")

	paths := GetPaths()
	return parser.ParseFiles([]string{targetFileName}, paths)
}

func GetPaths() []string {
	// TODO: to intelligence
	path := os.Getenv("PROTO_PATH") +
		":" + os.Getenv("GOPATH") + "/src/github.com/protocolbuffers/protobuf/src" +
		":" + targetDir

	paths := strings.Split(path, ":")

	return paths
}
