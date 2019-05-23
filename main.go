package main

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/NoahOrberg/protobuf_langserver/langserver"
	"github.com/sourcegraph/jsonrpc2"
)

func main() {
	if err := run(); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}

func run() error {
	lf := os.Getenv("HOME") + "/" + "p.log"
	logfile := &lf
	var logW io.Writer
	if *logfile == "" {
		logW = os.Stderr
	} else {
		f, err := os.Create(*logfile)
		if err != nil {
			return err
		}
		defer f.Close()
		logW = io.MultiWriter(os.Stderr, f)
	}

	log.SetOutput(logW)
	newHandler := func() (jsonrpc2.Handler, io.Closer) {
		return langserver.NewHandler(), ioutil.NopCloser(strings.NewReader(""))
	}

	var connOpt []jsonrpc2.ConnOpt

	log.Println("protobuf_langserver: reading on stdin, writing on stdout")
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
		log.Println(err)
	}
	log.Println("connection closed")

	return nil
}
