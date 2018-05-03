package main

import (
	"fmt"
	"net/url"
	"strings"
)

type parsedPath struct {
	rName  string
	id     string
	params *url.Values
}

// parsePath splits request path into resource name, id, and parameters and returns parsedPath struct.
// Request path example: /resource/id?params
func parsePath(requestPath string) (*parsedPath, error) {
	p := &parsedPath{}
	var path string
	if strings.Contains(requestPath, "?") {
		s := strings.Split(requestPath, "?")
		path = s[0]

		params, err := url.ParseQuery(s[1])
		if err != nil {
			return nil, err
		}
		p.params = &params
	} else {
		path = requestPath
		p.params = nil
	}

	// ignore empty strings and take two values
	names := make([]string, 0)
	for _, v := range strings.Split(path, "/") {
		if v != "" {
			names = append(names, v)
		}
	}
	if len(names) >= 2 {
		p.rName = names[0]
		p.id = strings.Join(names[1:], "/")
	} else if len(names) == 1 {
		p.rName = names[0]
		p.id = ""
	} else {
		return nil, fmt.Errorf("request path %s has resource name", requestPath)
	}

	return p, nil
}
