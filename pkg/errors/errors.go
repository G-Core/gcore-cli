package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/G-core/cli/pkg/output"
)

var ErrAborted = errors.New("operation aborted")

func ParseCloudErr(body []byte) *CliError {
	s := struct {
		Message string `json:"message"`
	}{}

	if err := json.Unmarshal(body, &s); err != nil {
		log.Println(err)
		output.Print(err)

		return nil
	}

	return &CliError{
		Err:  fmt.Errorf("%s", s.Message),
		Code: 1,
	}
}
