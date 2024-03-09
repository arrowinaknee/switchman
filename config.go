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
	server = &ServerConfig{}

	tokens.ReadExact("{")
	for {
		var token = tokens.ReadNext()
		if token == "}" {
			break
		} else if !token.IsLiteral() {
			log.Fatalf("Unexpected %s, server property name or '}' expected", token.Quote())
		}
		switch token {
		case "locations":
			server.endpoints = readLocations(tokens)
		default:
			log.Fatalf("%s is not a valid server property", token.Quote())
		}
	}

	return
}

func readLocations(tokens *tokenReader) (locations []Endpoint) {
	tokens.ReadExact("{")
	for {
		var token = tokens.ReadNext()
		if token == "}" {
			break
		} else if !token.IsLiteral() {
			log.Fatalf("Unexpected %s, location path or '}' was expected", token.Quote())
		}
		var endpoint Endpoint
		endpoint.location = token.String()
		tokens.ReadExact(":")

		var ep_type = tokens.ReadLiteral()
		switch ep_type {
		case "files":
			endpoint.function = readEpFiles(tokens)
		case "redirect":
			endpoint.function = readEpRedirect(tokens)
		default:
			log.Fatalf("%s is not a valid endpoint type", token.Quote())
		}
		locations = append(locations, endpoint)
	}
	return
}

func readEpFiles(tokens *tokenReader) (fun *EndpointFiles) {
	fun = &EndpointFiles{}

	tokens.ReadExact("{")
	for {
		var token = tokens.ReadNext()
		if token == "}" {
			break
		} else if !token.IsLiteral() {
			log.Fatalf("Unexpected %s, property name or '}' was expected", token.Quote())
		}
		switch token {
		case "sources":
			tokens.ReadExact(":")
			fun.fileRoot = tokens.ReadLiteral().String()
		default:
			log.Fatalf("%s is not a valid files endpoint property", token.Quote())
		}
	}
	return
}

func readEpRedirect(tokens *tokenReader) (fun *EndpointRedirect) {
	fun = &EndpointRedirect{}

	tokens.ReadExact("{")
	for {
		var token = tokens.ReadNext()
		if token == "}" {
			break
		} else if !token.IsLiteral() {
			log.Fatalf("Unexpected %s, property name or '}' was expected", token.Quote())
		}
		switch token {
		case "target":
			tokens.ReadExact(":")
			fun.target = tokens.ReadLiteral().String()
		default:
			log.Fatalf("%s is not a valid redirect endpoint property", token.Quote())
		}
	}
	return
}
