package task

import (
	v "github.com/go-ozzo/ozzo-validation"
)

type Task struct {
	Method  string            `json:"method"`
	Address string            `json:"address"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func (t *Task) Validate() error {
	return v.ValidateStruct(t,
		v.Field(&t.Method, v.Required, v.NotNil),
		v.Field(&t.Address, v.Required, v.NotNil),
	)
}
