package service

import (
	"github.com/pkg/errors"
	"megin/library/context/api"
	"megin/library/errs"
	"megin/library/logger"
)

type Service struct {
	ctx     *api.Context
	TraceId string
	Log     logger.Log
}

func (s *Service) initialize(ctx *api.Context) *Service {
	s.TraceId = ctx.TraceId
	s.Log = ctx.Log
	s.ctx = ctx
	return s
}

func (s *Service) errorMessage(message string, code ...int32) error {
	if len(code) > 0 {
		return errs.NewBusinessError(code[0], message)
	}
	return errs.NewBusinessError(500, message)
}

func (s *Service) errorCodeMessage(code int32, message string) error {
	return errs.NewBusinessError(code, message)
}

func (s *Service) error(err error, message ...string) error {
	if err == nil {
		return err
	}

	if len(message) > 0 {
		return errs.NewNormalError(500, message[0], errors.WithStack(err))
	}

	return errs.NewNormalError(500, "服务器错误", errors.WithStack(err))
}

func (s *Service) errorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return err
	}
	return errors.Wrapf(err, format, args)
}
