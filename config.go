package main

import (
	"io"
	"log"
)

func ParseServerConfig(source io.Reader) *ServerConfig {
	tokens, err := NewTokenReader(source)
	if err != nil {
		log.Fatal(err)
	}

	return readConfig(tokens)
}

func readConfig(tokens *tokenReader) (server *ServerConfig) {
	tokens.ReadExact("server")
	server = readServer(tokens)
	tokens.ReadExact(EOF)
	return
}

func readServer(tokens *tokenReader) (server *ServerConfig) {
	/*server {
		locations: {...}
	}*/
	server = &ServerConfig{}

	tokens.ReadStruct(func(tokens *tokenReader, field token) {
		switch field {
		case "locations":
			server.endpoints = readLocations(tokens)
		default:
			log.Fatalf("%s is not a valid server property", field.Quote())
		}
	})
	return
}

func readLocations(tokens *tokenReader) (locations []Endpoint) {
	/*locations{
		path: endpoint_type {...}
		...: ...
	}*/
	tokens.ReadStruct(func(tokens *tokenReader, field token) {
		var endpoint Endpoint

		endpoint.location = field.String()

		var ep_type = tokens.ReadProperty()
		switch ep_type {
		case "files":
			endpoint.function = readEpFiles(tokens)
		case "redirect":
			endpoint.function = readEpRedirect(tokens)
		default:
			log.Fatalf("%s is not a valid endpoint type", ep_type.Quote())
		}
		locations = append(locations, endpoint)
	})
	return
}

func readEpFiles(tokens *tokenReader) (fun *EndpointFiles) {
	/*files {
		sources: path
	}*/
	fun = &EndpointFiles{}

	tokens.ReadStruct(func(tokens *tokenReader, field token) {
		switch field {
		case "sources":
			fun.fileRoot = tokens.ReadProperty().String()
		default:
			log.Fatalf("%s is not a valid files endpoint property", field.Quote())
		}
	})
	return
}

func readEpRedirect(tokens *tokenReader) (fun *EndpointRedirect) {
	/*redirect = {
		target: path
	}*/
	fun = &EndpointRedirect{}

	tokens.ReadStruct(func(tokens *tokenReader, field token) {
		switch field {
		case "target":
			fun.target = tokens.ReadProperty().String()
		default:
			log.Fatalf("%s is not a valid redirect endpoint property", field.Quote())
		}
	})
	return
}
