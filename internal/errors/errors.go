package errors

import (
	"encoding/json"
	"errors"
	"fmt"
)

var ErrAborted = errors.New("operation aborted")

// TODO: It's not CLI work. SDK should do it.
func ParseCloudErr(body []byte) *CliError {
	s := struct {
		Message string `json:"message"`
	}{}

	if err := json.Unmarshal(body, &s); err != nil {
		return nil
	}

	return &CliError{
		Err:  fmt.Errorf("%s", s.Message),
		Code: 1,
	}
}
