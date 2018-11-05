// Copyright 2018 Adam Shannon
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/adamdecaf/godepnotify/pkg/modfetch"
	"github.com/adamdecaf/godepnotify/pkg/modparse"
	"github.com/adamdecaf/godepnotify/pkg/relparse"

	"github.com/gorilla/mux"
	moovhttp "github.com/moov-io/base/http"
)

func addScrapeEndpoint(r *mux.Router) {
	r.Methods("POST").Path("/scrape").HandlerFunc(scrapeEndpoint)
}

func getImportPath(r *http.Request) string {
	return r.URL.Query().Get("importPath")
}

type module struct {
	ImportPath string `json:"importPath"`
	Version    string `json:"version"`
}

func scrapeEndpoint(w http.ResponseWriter, r *http.Request) {
	importPath := getImportPath(r)
	if importPath == "" {
		moovhttp.Problem(w, errors.New("missing import path"))
		return
	}

	// Grab repo
	f, err := modfetch.New(importPath)
	if err != nil {
		moovhttp.Problem(w, fmt.Errorf("problem grabbing %s: %v", importPath, err))
		return
	}
	dir, err := f.Load()
	if err != nil {
		moovhttp.Problem(w, fmt.Errorf("problem loading %s: %v", importPath, err))
		return
	}

	// Find Modules
	mods, err := modparse.ParseFile(filepath.Join(dir, "go.sum"))
	if err != nil {
		moovhttp.Problem(w, fmt.Errorf("problem parsing %s go.sum: %v", importPath, err))
		return
	}

	// Render json
	var modules []module
	mods.ForEach(func(path string, ver *modparse.Version) {
		modules = append(modules, module{
			ImportPath: path,
			Version:    ver.String(),
		})
	})
	if err = json.NewEncoder(w).Encode(modules); err != nil {
		moovhttp.Problem(w, err)
		return
	}

	relparse.Parse(nil)
}