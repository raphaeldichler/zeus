// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package server

import (
	"fmt"
	"net/http"
)

func filterByMethod(method string, controllers []Controller) *Controller {
	for _, ctr := range controllers {
		if ctr.Method == method {
			return &Controller{
				Method: ctr.Method,
				Path:   ctr.Path,
				Run:    ctr.Run,
			}
		}
	}

	return nil
}

func (self *ServerConfig) setupControllers(mux *http.ServeMux) {
	grouped := make(map[string][]Controller)
	for _, ctr := range self.Controllers {
		group, ok := grouped[ctr.Path]
		if !ok {
			group = make([]Controller, 0)
		}

		grouped[ctr.Path] = append(group, ctr)
	}

	for path, ctr := range grouped {
		post := filterByMethod("POST", ctr)
		get := filterByMethod("GET", ctr)
		put := filterByMethod("PUT", ctr)
		del := filterByMethod("DELETE", ctr)

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.Method)
			switch r.Method {
			case "POST":
				if post == nil {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				post.Run(w, r)

			case "GET":
				if get == nil {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				get.Run(w, r)

			case "PUT":
				if put == nil {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				put.Run(w, r)

			case "DELETE":
				if del == nil {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				del.Run(w, r)

			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})
	}

}
