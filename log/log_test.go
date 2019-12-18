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
		expected string
	}{
		{
			f: func() {
				log.L().Info("Info")
			},
			expected: "Info",
		},
		{
			f: func() {
				log.L().Warn("Warn")
			},
			expected: "Warn",
		},
		{
			f: func() {
				log.L().Error("Error")
			},
			expected: "Error",
		},
		{
			f: func() {
				log.L().Fatal("Fatal")
			},
			expected: "Fatal",
		},
	}

	for _, tc := range tcs {
		buf := new(bytes.Buffer)

		log.Init(buf)
		tc.f()

		actual, err := ioutil.ReadAll(buf)
		if err != nil {
			t.Fatalf("expected err is nil, buf got %s", err)
		}

		println(string(actual))

		if strings.HasSuffix(string(actual), tc.expected) {
			t.Fatalf("expected %s, but got %s", tc.expected, actual)
		}
	}
}
