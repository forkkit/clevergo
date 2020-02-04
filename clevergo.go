// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Package clevergo is a trie based high performance HTTP request router.
//
// A trivial example is:
//
//  package main
//
//  import (
//      "fmt"
//      "log"
//      "net/http"
//
//      "github.com/clevergo/clevergo"
//  )
//
//  func Index(w http.ResponseWriter, r *http.Request) {
//      fmt.Fprint(w, "Welcome!\n")
//  }
//
//  func Hello(w http.ResponseWriter, r *http.Request) {
//      ps := clevergo.GetParams(r)
//      fmt.Fprintf(w, "hello, %s!\n", ps.String("name"))
//  }
//
//  func main() {
//      app := clevergo.New()
//      app.Get("/", Index)
//      app.Get("/hello/:name", Hello)
//
//      log.Fatal(http.ListenAndServe(":8080", router))
//  }
//
// The router matches incoming requests by the request method and the path.
// If a handle is registered for this path and method, the router delegates the
// request to that function.
// For the methods Get, Post, Put, Patch, Delete and Options shortcut functions exist to
// register handles, for all other methods router.Handle can be used.
//
// The registered path, against which the router matches incoming requests, can
// contain two types of parameters:
//  Syntax    Type
//  :name     named parameter
//  *name     catch-all parameter
//
// Named parameters are dynamic path segments. They match anything until the
// next '/' or the path end:
//  Path: /blog/:category/:post
//
//  Requests:
//   /blog/go/request-routers            match: category="go", post="request-routers"
//   /blog/go/request-routers/           no match, but the router would redirect
//   /blog/go/                           no match
//   /blog/go/request-routers/comments   no match
//
// Catch-all parameters match anything until the path end, including the
// directory index (the '/' before the catch-all). Since they match anything
// until the end, catch-all parameters must always be the final path element.
//  Path: /files/*filepath
//
//  Requests:
//   /files/                             match: filepath="/"
//   /files/LICENSE                      match: filepath="/LICENSE"
//   /files/templates/article.html       match: filepath="/templates/article.html"
//   /files                              no match, but the router would redirect
//
// The value of parameters is saved as a slice of the Param struct, consisting
// each of a key and a value. The slice is passed to the Handle func as a third
// parameter.
// There are two ways to retrieve the value of a parameter:
//  // by the name of the parameter
//  ps := GetParams(req) // retrieves params of the given request
//  user := ps.String("user") // defined by :user or *user
//
//  // by the index of the parameter. This way you can also get the name (key)
//  thirdKey   := ps[2].Key   // the name of the 3rd parameter
//  thirdValue := ps[2].Value // the value of the 3rd parameter
package clevergo

import (
	"net"
	"net/http"
)

// Application application is a wrapper of Router and http.Server.
type Application struct {
	*Router
	*http.Server

	middlewares []Middleware
	onCleanUp   []func()
}

// New returns an application.
func New(addr string) *Application {
	return &Application{
		Server: &http.Server{
			Addr: addr,
		},
		Router: NewRouter(),
	}
}

// Use registers middlewares.
func (app *Application) Use(middlewares ...Middleware) {
	app.middlewares = append(app.middlewares, middlewares...)
}

func (app *Application) prepare() {
	app.Server.Handler = Chain(app.Router, app.middlewares...)
}

// ListenAndServe overrides http.Server.ListenAndServe with extra preparations.
func (app *Application) ListenAndServe() error {
	app.prepare()
	return app.Server.ListenAndServe()
}

// ListenAndServeTLS overrides http.Server.ListenAndServeTLS with extra preparations.
func (app *Application) ListenAndServeTLS(certFile, keyFile string) error {
	app.prepare()
	return app.Server.ListenAndServeTLS(certFile, keyFile)
}

// ListenAndServeUnix listens on the Unix socket app.Server.Addr
// and then calls Serve to handle requests on incoming connections.
func (app *Application) ListenAndServeUnix() error {
	l, err := net.Listen("unix", app.Addr)
	if err != nil {
		return err
	}
	return app.Serve(l)
}

// Serve overrides http.Server.Serve with extra preparations.
func (app *Application) Serve(l net.Listener) error {
	app.prepare()
	return app.Server.Serve(l)
}

// ServeTLS overrides http.Server.ServeTLS with extra preparations.
func (app *Application) ServeTLS(l net.Listener, certFile, keyFile string) error {
	app.prepare()
	return app.Server.ServeTLS(l, certFile, keyFile)
}

// RegisterOnCleanUp registers a function to call on CleanUp.
func (app *Application) RegisterOnCleanUp(fs func()) {
	app.onCleanUp = append(app.onCleanUp, fs)
}

// CleanUp calls clean up functions before closing server.
func (app *Application) CleanUp() {
	for _, f := range app.onCleanUp {
		f()
	}
}
