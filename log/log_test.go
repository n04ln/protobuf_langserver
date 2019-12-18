package log_test

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/NoahOrberg/protobuf_langserver/log"
)

func TestInit(t *testing.T) {
	tcs := []struct {
		f        func()
		name     string
		expected string
	}{
		{
			f: func() {
				log.L().Info("some data")
			},
			name:     "normal case",
			expected: "\tinfo\tsome data",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			log.Init(buf)
			tc.f()

			actual, err := ioutil.ReadAll(buf)
			if err != nil {
				t.Fatalf("expected err is nil, buf got %s", err)
			}

			if strings.HasSuffix(string(actual), tc.expected) {
				t.Fatalf("expected %s, but got %s", tc.expected, actual)
			}
		})
	}
}
