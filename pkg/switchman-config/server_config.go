package main

import (
	"io"
)

func ParseServerConfig(source io.Reader) (*ServerConfig, error) {
	return readConfig(NewConfigReader(source))
}

func readConfig(config *ConfigReader) (server *ServerConfig, err error) {
	if err = config.ReadExact("server"); err != nil {
		return
	}
	server, err = readServer(config)
	if err != nil {
		return
	}
	if err = config.ReadExact(EOF); err != nil {
		server = nil
	}
	return
}

func readServer(config *ConfigReader) (server *ServerConfig, err error) {
	/*server {
		locations: {...}
	}*/
	server = &ServerConfig{}

	err = config.ReadStruct(func(config *ConfigReader, field token) (err error) {
		switch field {
		case "locations":
			server.endpoints, err = readLocations(config)
		default:
			err = errUnrecognized(field, "recognized server property")
		}
		return
	})
	if err != nil {
		return nil, err
	}
	return
}

func readLocations(config *ConfigReader) (locations []Endpoint, err error) {
	/*locations{
		path: endpoint_type {...}
		...: ...
	}*/
	err = config.ReadStruct(func(config *ConfigReader, field token) (err error) {
		var endpoint Endpoint

		endpoint.location = field.String()

		var ep_type token
		ep_type, err = config.ReadProperty()
		if err != nil {
			return
		}
		switch ep_type {
		case "files":
			endpoint.function, err = readEpFiles(config)
		case "redirect":
			endpoint.function, err = readEpRedirect(config)
		default:
			return errUnrecognized(ep_type, "recognized endpoint type")
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

func readEpFiles(config *ConfigReader) (fun *EndpointFiles, err error) {
	/*files {
		sources: path
	}*/
	fun = &EndpointFiles{}

	err = config.ReadStruct(func(config *ConfigReader, field token) (err error) {
		switch field {
		case "sources":
			fun.fileRoot, err = config.ReadPropertyName()
		default:
			err = errUnrecognized(field, "recognized files endpoint property")
		}
		return
	})
	return
}

func readEpRedirect(config *ConfigReader) (fun *EndpointRedirect, err error) {
	/*redirect = {
		target: path
	}*/
	fun = &EndpointRedirect{}

	err = config.ReadStruct(func(config *ConfigReader, field token) (err error) {
		switch field {
		case "target":
			fun.target, err = config.ReadPropertyName()
		default:
			err = errUnrecognized(field, "recognized redirect endpoint property")
		}
		return
	})
	if err != nil {
		return nil, err
	}
	return
}
