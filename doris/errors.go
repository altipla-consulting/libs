package doris

import (
	"github.com/bufbuild/connect-go"

	"libs.altipla.consulting/errors"
)

func Errorf(code connect.Code, msg string, args ...any) error {
	return errors.Trace(connect.NewError(code, errors.Errorf(msg, args...)))
}
