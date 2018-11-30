package presenters

type ClientResponse struct {
	Key           string              `json:"key"`
	HTTPStatus    string              `json:"httpStatus,omitempty"`
	Headers       map[string][]string `json:"headers,omitempty"`
	ContentLength int64               `json:"contentLength,omitempty"`
	Done          bool                `json:"done"`
}
