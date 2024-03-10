package main

import (
	"reflect"
	"strings"
	"testing"
)

type configCase struct {
	name   string
	source string
	result *ServerConfig
	err    error
}

var basic_case = configCase{
	name: "Basic",
	source: `
server {
	locations {
		/test: files {
			sources: /test/
		}
		/redirect: redirect {
			target: /test
		}
	}
}`,
	result: &ServerConfig{
		endpoints: []Endpoint{
			{location: "/test", function: &EndpointFiles{"/test/"}},
			{location: "/redirect", function: &EndpointRedirect{"/test"}},
		},
	},
}

func TestParseServerConfig(t *testing.T) {
	// TODO: proper error checking
	tests := []configCase{
		basic_case,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := ParseServerConfig(strings.NewReader(tt.source)); !reflect.DeepEqual(got, tt.result) || err != tt.err {
				t.Errorf("ParseServerConfig() = %v, want %v", got, tt.result)
			}
		})
	}
}
