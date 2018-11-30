package api

import (
	"encoding/json"
	"io"
	"net/http"
)

type EmptySuccess map[string]string

type Response struct {
	Success  bool        `json:"success"`
	Errors   interface{} `json:"errors"`
	Data     interface{} `json:"data"`
	HttpCode int         `json:"-"`
}

func ErrorResponse(errors interface{}) *Response {
	return &Response{
		Success:  false,
		Errors:   errors,
		Data:     nil,
		HttpCode: http.StatusBadRequest,
	}
}

func ServerErrorResponse(serverErrorMessage string, httpCode int) *Response {
	return &Response{
		Success:  false,
		Errors:   map[string]string{"serverError": serverErrorMessage},
		Data:     nil,
		HttpCode: httpCode,
	}
}

func InternalServerError() *Response {
	return ServerErrorResponse("Internal server error", http.StatusInternalServerError)
}

func NoContentResponse() *Response {
	return ServerErrorResponse("No content", http.StatusNoContent)
}

func SuccessResponse(data interface{}) *Response {
	return &Response{
		Success:  true,
		Errors:   map[string]error{},
		Data:     data,
		HttpCode: http.StatusOK,
	}
}

func EmptySuccessResponse() *Response {
	return SuccessResponse(EmptySuccess{})
}

func ToJSON(w io.Writer, response Response) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	return enc.Encode(response)
}
