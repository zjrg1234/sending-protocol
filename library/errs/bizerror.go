package errs

import "fmt"

const DS = "###"

// 业务错误,提示类错误,非程序异常
type BusinessError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func NewBusinessError(code int32, message string) *BusinessError {
	return &BusinessError{code, message}
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("%v%s%v", e.Code, DS, e.Message)
}

func IsBusinessError(err error) (*BusinessError, bool) {
	if err != nil {
		switch err.(type) {
		case *BusinessError:
			return err.(*BusinessError), true
		default:
			return nil, false
		}
	}
	return nil, false
}
