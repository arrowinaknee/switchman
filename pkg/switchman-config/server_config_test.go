package appconfig

import (
	"reflect"
	"strings"
	"testing"

	"github.com/arrowinaknee/switchman/pkg/servers/http"
)

func TestParseServerConfig(t *testing.T) {
	type testCase struct {
		name    string
		source  string
		result  *http.Server
		wantErr bool
	}

	tests := []testCase{
		{
			name: "full_config",
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
			result: &http.Server{
				Endpoints: []http.Endpoint{
					{Location: "/test", Function: &http.EndpointFiles{FileRoot: "/test/"}},
					{Location: "/redirect", Function: &http.EndpointRedirect{Target: "/test"}},
				},
			},
			wantErr: false,
		}, {
			name: "missing_parameter",
			source: `
			server {
				locations {
					/test: files {
						sources: 
					}
				}
			}`,
			wantErr: true,
		}, {
			name: "missing_parenthesis",
			source: `
			server {
				locations {
					/test: files {
						sources: 
				}
			}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseServer(strings.NewReader(tt.source))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseServerConfig() error = \"%v\", wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.result) {
				t.Errorf("ParseServerConfig() = %v, want %v", got, tt.result)
			}
		})
	}
}
