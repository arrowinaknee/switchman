package appconfig

import (
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/arrowinaknee/switchman/pkg/config"
	"github.com/arrowinaknee/switchman/pkg/servers/http"
)

var urlRegexp = regexp.MustCompile(`^(?:(?P<proto>[a-zA-Z0-9]+)://)?(?P<hostname>[0-9a-zA-Z\-\.]+)?(?P<port>:[0-9]+)?(?P<path>/[0-9a-zA-Z\-\._/%&]*)?(?P<query>\?.*)?(?P<fragment>#.*)?$`)
var hostRegexp = regexp.MustCompile(`^(([a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]|[a-zA-Z0-9])\.?)+$`)

func ParseServer(source io.Reader) (*http.Server, error) {
	return readConfig(config.NewReader(source))
}

func readConfig(conf *config.Reader) (server *http.Server, err error) {
	if err = conf.ReadExact("server"); err != nil {
		return
	}
	server, err = readServer(conf)
	if err != nil {
		return
	}
	if err = conf.ReadExact(config.EOF); err != nil {
		server = nil
	}
	return
}

func readServer(conf *config.Reader) (server *http.Server, err error) {
	/*server {
		locations: {...}
	}*/
	server = &http.Server{}

	err = conf.ReadStruct(func(conf *config.Reader, field config.Token) (err error) {
		switch field {
		case "endpoints":
			server.Endpoints, err = readEndpoints(conf)
		default:
			err = conf.ErrUnrecognized("server property")
		}
		return
	})
	if err != nil {
		return nil, err
	}
	return
}

func readEndpoints(conf *config.Reader) (locations []http.Endpoint, err error) {
	/*locations{
		path: endpoint_type {...}
		...: ...
	}*/
	err = conf.ReadStruct(func(conf *config.Reader, field config.Token) (err error) {
		var endpoint http.Endpoint

		var t config.Token
		t, err = field.Unescaped()
		if err != nil {
			return
		}
		endpoint.Location = t.String()

		var ep_type config.Token
		err = conf.ReadSeparator()
		if err != nil {
			return
		}
		ep_type, err = conf.ReadName()
		if err != nil {
			return
		}
		switch ep_type {
		case "files":
			endpoint.Function, err = readEpFiles(conf)
		case "redirect":
			endpoint.Function, err = readEpRedirect(conf)
		case "proxy":
			endpoint.Function, err = readEpProxy(conf)
		default:
			return conf.ErrUnrecognized("endpoint type")
		}
		if err != nil {
			return
		}
		locations = append(locations, endpoint)
		return
	})
	if err != nil {
		return nil, err
	}
	return
}

func readEpFiles(conf *config.Reader) (fun *http.EndpointFiles, err error) {
	/*files {
		sources: path
	}*/
	fun = &http.EndpointFiles{}

	err = conf.ReadStruct(func(conf *config.Reader, field config.Token) (err error) {
		err = conf.ReadSeparator()
		if err != nil {
			return
		}
		var t config.Token
		switch field {
		case "sources":
			t, err = conf.ReadString()
			if err != nil {
				return
			}
			t, err = t.Unescaped()
			if err != nil {
				return err
			}
			fun.Source = t.String()
		default:
			err = conf.ErrUnrecognized("files endpoint property")
		}
		return
	})
	if err != nil {
		return nil, err
	}
	return
}

func readEpRedirect(conf *config.Reader) (fun *http.EndpointRedirect, err error) {
	/*redirect {
		url: path
	}*/
	fun = &http.EndpointRedirect{}

	err = conf.ReadStruct(func(conf *config.Reader, field config.Token) (err error) {
		err = conf.ReadSeparator()
		if err != nil {
			return
		}
		var t config.Token
		switch field {
		case "url":
			t, err = conf.ReadString()
			if err != nil {
				return
			}
			t, err = t.Unescaped()
			if err != nil {
				return err
			}
			fun.URL = t.String()
		default:
			err = conf.ErrUnrecognized("redirect endpoint property")
		}
		return
	})
	if err != nil {
		return nil, err
	}
	return
}

func readEpProxy(conf *config.Reader) (fun *http.EndpointProxy, err error) {
	/*proxy {
		url: "http://example.com:8080/hello"
	}*/
	fun = &http.EndpointProxy{
		Proto: "http",
		Host:  "localhost:80",
		Path:  "/",
	}

	err = conf.ReadStruct(func(conf *config.Reader, field config.Token) (err error) {
		err = conf.ReadSeparator()
		if err != nil {
			return
		}
		var t config.Token
		switch field {
		case "url":
			const url_err_text = "malformed or unsupported url"
			t, err = conf.ReadString()
			if err != nil {
				return
			}
			t, err = t.Unescaped()
			if err != nil {
				return err
			}
			url := t.String()

			if len(url) == 0 {
				return conf.Errorf("%s, expected url format: [http://][hostname][:port][/path]", url_err_text)
			}
			match := urlRegexp.FindStringSubmatch(url)
			if len(match) == 0 {
				return conf.Errorf("%s, expected url format: [http://][hostname][:port][/path]", url_err_text)
			}
			result := make(map[string]string)
			for i, name := range urlRegexp.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = match[i]
				}
			}

			proto := ""
			host := ""
			port := ""
			path := ""

			if proto = result["proto"]; proto != "" {
				if proto != "http" {
					return conf.Errorf("%s: %s protocol is not supported", url_err_text, proto)
				}
				fun.Proto = proto
			}
			if host = result["hostname"]; host != "" {
				if !hostRegexp.MatchString(host) {
					return conf.Errorf("%s: invalid hostname %s", url_err_text, host)
				}
			}
			if port = result["port"]; port != "" {
				portn, err := strconv.Atoi(port[1:])
				if err != nil || portn > 65535 {
					return fmt.Errorf("%s: port must be a number from 1 to 65535", url_err_text)
				}
			}
			path = result["path"]

			if query := result["query"]; query != "" {
				return fmt.Errorf("%s: query is not allowed in proxy url", url_err_text)
			}
			if fragment := result["fragment"]; fragment != "" {
				return fmt.Errorf("%s: fragment is not allowed in proxy url", url_err_text)
			}

			if proto != "" {
				fun.Proto = proto
			}
			if host != "" || port != "" {
				if host == "" {
					host = "localhost"
				}
				if port == "" {
					port = ":80"
				}
				fun.Host = host + port
			}
			if path != "" {
				fun.Path = path
			}
		default:
			err = conf.ErrUnrecognized("proxy endpoint property")
		}
		return
	})
	if err != nil {
		return nil, err
	}
	return
}
