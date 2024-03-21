package appconfig

import (
	"reflect"
	"strings"
	"testing"

	"github.com/arrowinaknee/switchman/pkg/config"
	"github.com/arrowinaknee/switchman/pkg/servers/http"
)

func TestParseServer(t *testing.T) {
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
				endpoints {
					# comment line
					/test: files {
						sources: "E:/test website/"
					}
					"/redirect": redirect {
						target: /test
					}
				}
			}`,
			result: &http.Server{
				Endpoints: []http.Endpoint{
					{Location: "/test", Function: &http.EndpointFiles{FileRoot: "E:/test website/"}},
					{Location: "/redirect", Function: &http.EndpointRedirect{Target: "/test"}},
				},
			},
			wantErr: false,
		}, {
			name: "missing_parameter",
			source: `
			server {
				endpoints {
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
				endpoints {
					/test: files {
						sources: 
				}
			}`,
			wantErr: true,
		}, {
			name: "unexpected_string",
			source: `
			server {
				"endpoints" {
				}
			}`,
			wantErr: true,
		}, {
			name:    "empty",
			source:  "",
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

func Test_readServer(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *http.Server
		wantErr bool
	}{
		{
			name: "full",
			input: `{
				endpoints {}
			}`,
			want: &http.Server{
				Endpoints: nil,
			},
			wantErr: false,
		}, {
			name:    "empty",
			input:   `{}`,
			want:    &http.Server{},
			wantErr: false,
		}, {
			name: "wrong_property",
			input: `{
				locations {}
			}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := config.NewReader(strings.NewReader(tt.input))
			got, err := readServer(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("readServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readEndpoints(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []http.Endpoint
		wantErr bool
	}{
		{
			name: "all",
			input: `{
				/files: files {
					sources: "E:/files/"
				}
				/redirect: redirect {
					target: example.com/redirect
				}
			}`,
			want: []http.Endpoint{
				{
					Location: "/files",
					Function: &http.EndpointFiles{
						FileRoot: "E:/files/",
					},
				}, {
					Location: "/redirect",
					Function: &http.EndpointRedirect{
						Target: "example.com/redirect",
					},
				},
			},
			wantErr: false,
		}, {
			name:    "empty",
			input:   "{}",
			want:    nil,
			wantErr: false,
		}, {
			name: "no_location",
			input: `{
				redirect {
					target: /
				}
			}`,
			wantErr: true,
		}, {
			name: "no_type",
			input: `{
				/test: {
					target: /
				}
			}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := config.NewReader(strings.NewReader(tt.input))
			got, err := readEndpoints(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("readEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readEndpoints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readEpFiles(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *http.EndpointFiles
		wantErr bool
	}{
		{
			name: "full",
			input: `{
				sources: "E:/test/"
			}`,
			want: &http.EndpointFiles{
				FileRoot: "E:/test/",
			},
			wantErr: false,
		}, {
			name:    "empty",
			input:   "{}",
			want:    &http.EndpointFiles{},
			wantErr: false,
		}, {
			name: "wrong_property",
			input: `{
				target: /
			}`,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := config.NewReader(strings.NewReader(tt.input))
			got, err := readEpFiles(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("readEpFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readEpFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readEpRedirect(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *http.EndpointRedirect
		wantErr bool
	}{
		{
			name: "full",
			input: `{
				target: /
			}`,
			want: &http.EndpointRedirect{
				Target: "/",
			},
			wantErr: false,
		}, {
			name:    "empty",
			input:   "{}",
			want:    &http.EndpointRedirect{},
			wantErr: false,
		}, {
			name: "wrong_property",
			input: `{
				sources: /
			}`,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := config.NewReader(strings.NewReader(tt.input))
			got, err := readEpRedirect(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("readEpRedirect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readEpRedirect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkParseServer(b *testing.B) {
	input := `
	server {
		endpoints {
			/test1: files {
				sources: "E:/test website/"
			}
			/test2: files {
				sources: tests/test
			}
			/redirect: redirect {
				target: /test1
			}
		}
	}`
	for i := 0; i < b.N; i++ {
		_, err := ParseServer(strings.NewReader(input))
		if err != nil {
			b.Error(err)
		}
	}
}
