package appconfig

import (
	"io"

	"github.com/arrowinaknee/switchman/pkg/config"
	"github.com/arrowinaknee/switchman/pkg/servers/http"
)

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
		case "locations":
			server.Endpoints, err = readLocations(conf)
		default:
			err = config.ErrUnrecognized(field, "recognized server property")
		}
		return
	})
	if err != nil {
		return nil, err
	}
	return
}

func readLocations(conf *config.Reader) (locations []http.Endpoint, err error) {
	/*locations{
		path: endpoint_type {...}
		...: ...
	}*/
	err = conf.ReadStruct(func(conf *config.Reader, field config.Token) (err error) {
		var endpoint http.Endpoint

		endpoint.Location = field.String()

		var ep_type config.Token
		ep_type, err = conf.ReadProperty()
		if err != nil {
			return
		}
		switch ep_type {
		case "files":
			endpoint.Function, err = readEpFiles(conf)
		case "redirect":
			endpoint.Function, err = readEpRedirect(conf)
		default:
			return config.ErrUnrecognized(ep_type, "recognized endpoint type")
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
		switch field {
		case "sources":
			fun.FileRoot, err = conf.ReadPropertyName()
		default:
			err = config.ErrUnrecognized(field, "recognized files endpoint property")
		}
		return
	})
	return
}

func readEpRedirect(conf *config.Reader) (fun *http.EndpointRedirect, err error) {
	/*redirect = {
		target: path
	}*/
	fun = &http.EndpointRedirect{}

	err = conf.ReadStruct(func(conf *config.Reader, field config.Token) (err error) {
		switch field {
		case "target":
			fun.Target, err = conf.ReadPropertyName()
		default:
			err = config.ErrUnrecognized(field, "recognized redirect endpoint property")
		}
		return
	})
	if err != nil {
		return nil, err
	}
	return
}