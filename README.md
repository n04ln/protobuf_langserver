# protobuf_langserver

[![CircleCI](https://circleci.com/gh/NoahOrberg/protobuf_langserver.svg?style=svg)](https://circleci.com/gh/NoahOrberg/protobuf_langserver)

## Installation

``` sh
$ make install
```

## Usage

- plz write below code in your `.config/nvim/init.vim` if u use `vim-lsp`.

``` vim
if executable('protobuf_langserver')
    au User lsp_setup call lsp#register_server({
        \ 'name': 'protobuf_server',
        \ 'cmd': {server_info->['protobuf_langserver']},
        \ 'whitelist': ['proto'],
        \ })
endif
```
