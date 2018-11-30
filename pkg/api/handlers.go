package api

import "net/http"

type HandlerWithReturn func(w http.ResponseWriter, req *http.Request) (*Response, error)

func ResponseHandler(withReturn HandlerWithReturn) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		resp, err := withReturn(w, req)
		if err != nil {
			resp = InternalServerError()
		}

		if resp != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.HttpCode)
			ToJSON(w, *resp)
		}
	})
}
