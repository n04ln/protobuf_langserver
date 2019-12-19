# protobuf_langserver

[![CircleCI](https://circleci.com/gh/n04ln/protobuf_langserver.svg?style=svg)](https://circleci.com/gh/n04ln/protobuf_langserver)

## Installation

``` sh
$ go get github.com/n04ln/protobuf_langserver
```

## Usage

- plz set `$PROTO_PATH` environment variable. However the directory of opened file and `${GOPATH}/src/github.com/protocolbuffers/protobuf/src` is already added. This language server through `$PROTO_PATH`, search each directory from left to right in the list.

``` sh
# Example:
export PROTO_PATH=/path/to/proto:/path/to/proto2
```

- plz write below code in your setting file (maybe `~/.vimrc` or `~/.config/nvim/init.vim`) if u use `vim-lsp`.

``` vim
" Example:
if executable('protobuf_langserver')
    au User lsp_setup call lsp#register_server({
        \ 'name': 'protobuf_server',
        \ 'cmd': {server_info->['protobuf_langserver']},
        \ 'whitelist': ['proto'],
        \ })
endif
```

### For developer

- if you wanna to see logs, plz set such as below. you can see `/path/to/logfile`.

``` vim
" Example with log:
if executable('protobuf_langserver')
    au User lsp_setup call lsp#register_server({
        \ 'name': 'protobuf_langserver',
        \ 'cmd': {server_info->['protobuf_langserver', '-log', '/path/to/logfile']},
        \ 'whitelist': ['proto'],
        \ })
endif
```

