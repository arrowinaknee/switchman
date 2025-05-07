package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type apiResponse struct {
	Ok          bool   `json:"ok"`
	Result      any    `json:"result"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

type responseWriter struct {
	httpwriter  http.ResponseWriter
	request     *http.Request
	hasResult   bool
	hasError    bool
	errCode     int
	description string
	result      any
	headerSent  bool
}

func (w *responseWriter) Result(result any) {
	if w.hasResult {
		fmt.Printf("Logical error: result written twice for %q. New result ignored.", w.request.URL.Path)
		return
	}
	w.hasResult = true
	w.result = result
	w.errCode = http.StatusOK

	if !w.headerSent {
		w.Header().Set("Content-Type", "application/json")
		w.httpwriter.WriteHeader(http.StatusOK)
		w.headerSent = true
	}
}

func (w *responseWriter) Ok() {
	w.Result(nil)
}

func (w *responseWriter) ErrorText(status int, description string) {
	if w.hasError {
		fmt.Printf("Logical error: error written twice for %q. New error ignored.", w.request.URL.Path)
		return
	}
	if w.result != nil {
		fmt.Printf("Logical error: error written after response for %q. Error ignored.", w.request.URL.Path)
		return
	}

	w.errCode = status
	w.hasError = true
	w.description = description

	w.Header().Set("Content-Type", "application/json")
	w.httpwriter.WriteHeader(status)
	w.headerSent = true
}

func (w *responseWriter) Error(status int) {
	w.ErrorText(status, "")
}

func (w *responseWriter) Header() http.Header {
	return w.httpwriter.Header()
}

func (w *responseWriter) Commit() error {
	resp := apiResponse{
		Ok:          !w.hasError,
		Result:      w.result,
		ErrorCode:   w.errCode,
		Description: w.description,
	}

	err := json.NewEncoder(w.httpwriter).Encode(resp)
	if err != nil {
		fmt.Printf("Error writing response to %q: %v", w.request.URL.Path, err)
		return err
	}
	return nil
}

func decoded[T any](next func(*responseWriter, *http.Request, *T)) func(*responseWriter, *http.Request) {
	return func(w *responseWriter, r *http.Request) {
		val := new(T)
		err := json.NewDecoder(r.Body).Decode(val)
		if err != nil {
			w.ErrorText(http.StatusBadRequest, "Could not parse request body")
			return
		}

		next(w, r, val)
	}
}
