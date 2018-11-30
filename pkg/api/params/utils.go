package params

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	v "github.com/go-ozzo/ozzo-validation"
)

func MustDecodeParams(reader io.Reader, to interface{}) error {
	bodyDecoder := json.NewDecoder(reader)

	return bodyDecoder.Decode(&to)
}

func MustValidateParams(validatable v.Validatable) error {
	validationErrors := validatable.Validate()

	if err, ok := validationErrors.(v.InternalError); ok == true {
		return errors.New(fmt.Sprintf("params.MustValidateParams, unable to validate params: %s", err))
	}

	return validationErrors
}
