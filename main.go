package main

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
	stdlog "log"
	"os"
	"strings"

	"github.com/n04ln/protobuf_langserver/langserver"
	"github.com/n04ln/protobuf_langserver/log"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

var (
	logFile string
)

func init() {
	flag.StringVar(&logFile, "log", "", "specify log file path locaiton")
	flag.Parse()
}

func main() {
	if err := run(); err != nil {
		log.L().Fatal(err.Error())
		os.Exit(1)
	}
}

func run() error {
	var logW io.Writer
	if logFile == "" {
		logW = os.Stderr
	} else {
		f, err := os.Create(logFile)
		if err != nil {
			return err
		}
		defer f.Close()
		logW = io.MultiWriter(os.Stderr, f)
	}
	log.Init(logW)
	stdlog.SetOutput(logW)

	newHandler := func() (jsonrpc2.Handler, io.Closer) {
		return langserver.NewHandler(), ioutil.NopCloser(strings.NewReader(""))
	}

	var connOpt []jsonrpc2.ConnOpt

	log.L().Info("protobuf_langserver: reading on stdin, writing on stdout")
	handler, closer := newHandler()
	<-jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(
			stdrwc{},
			jsonrpc2.VSCodeObjectCodec{}),
		handler,
		connOpt...).
		DisconnectNotify()
	err := closer.Close()
	if err != nil {
		log.L().Fatal("failed closer.Close()", zap.Error(err))
	}
	log.L().Info("connection closed")

	return nil
}
