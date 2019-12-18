module github.com/NoahOrberg/protobuf_langserver

go 1.12

replace github.com/myitcv/x => myitcv.io v0.0.0-20190927111909-7837eed0ff8e

require (
	github.com/NoahOrberg/x v0.0.0-20190513145353-63b658caea3e
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/sourcegraph/go-lsp v0.0.0-20181119182933-0c7d621186c1
	github.com/sourcegraph/jsonrpc2 v0.0.0-20190106185902-35a74f039c6a
	go.uber.org/zap v1.13.0
)
