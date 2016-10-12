// Copyright 2016 Eric Wollesen <ericw at xmtp dot net>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpapi

import (
	"bytes"
	"flag"
	"html/template"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/spacemonkeygo/spacelog"
)

var (
	logger = spacelog.GetLogger()

	address = flag.String("httpapi.address", ":8080",
		"Address to listen on for the HTTP API")
)

type router struct {
	*mux.Router
}

func New() *router {
	r := &router{Router: mux.NewRouter()}
	r.HandleFunc("/", r.handleRoot)
	r.StrictSlash(true)
	return r
}

func (r *router) Serve() error {
	go func() {
		logger.Errore(http.ListenAndServe(*address, r))
	}()
	logger.Infof("HTTP API listening on %q", *address)
	return nil
}

func (r *router) handleRoot(w http.ResponseWriter, req *http.Request) {
	logger.Debugf("HTTP request received %+v", req)
	routes := []string{}
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pt, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		match, err := regexp.MatchString("^/[^/]+$", pt)
		if err != nil {
			return err
		}
		if match {
			routes = append(routes, pt)
		}
		return nil
	})
	t, err := template.New("root").Parse(`<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Zenbot HTTP API</title>
	</head>
	<body>
                <h1>Zenbot HTTP API</h1>
                <p>
                    Endpoints:
                    <ul>
                    {{range .Routes}}
                    <li><a href="{{.}}">{{.}}</a></li>
                    {{end}}
                    </ul>
                </p>
	</body>
</html>`)
	data := struct {
		Routes []string
	}{
		Routes: routes,
	}
	buf := bytes.NewBufferString("")
	err = t.Execute(buf, data)
	if err != nil {
		logger.Errore(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error generating HTML"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func (r *router) HandleFunc(path string,
	fn func(http.ResponseWriter, *http.Request)) {
	r.Router.HandleFunc(path, fn)
}

func (r *router) ForPath(path string) Router {
	return &router{Router: r.PathPrefix("/discord").Subrouter()}
}
