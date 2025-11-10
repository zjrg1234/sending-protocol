package errs

import (
	"fmt"
)

// 普通错误,程序异常
type NormalError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Err     error
}

func NewNormalError(code int32, message string, err error) *NormalError {
	return &NormalError{code, message, err}
}

func (e *NormalError) Error() string {
	return fmt.Sprintf("%v%s%v", e.Code, DS, e.Message)
}

func IsNormalError(err error) (*NormalError, bool) {
	if err != nil {
		switch err.(type) {
		case *NormalError:
			return err.(*NormalError), true
		default:
			return nil, false
		}
	}
	return nil, false
}
