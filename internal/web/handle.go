package web

import "net/http"

type HandlerFunc func(w http.ResponseWriter, req *http.Request) error

func (r *Responder) Handle(handler HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := handler(w, req)
		if err != nil {
			r.Error(w, req, err)
		}
	})
}
