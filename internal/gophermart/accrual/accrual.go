package accrual

import (
	"context"

	"github.com/4aleksei/gmart/internal/gophermart/service"

	"github.com/4aleksei/gmart/internal/common/logger"

	"github.com/4aleksei/gmart/internal/gophermart/config"
)

type (
	HandlersAccrual struct {
		cfg *config.Config
		l   *logger.ZapLogger
		s   *service.HandleService
	}
)

func NewAccrual(cfg *config.Config, s *service.HandleService, l *logger.ZapLogger) *HandlersAccrual {

	return &HandlersAccrual{
		cfg: cfg,
		l:   l,
		s:   s,
	}
}

func (a *HandlersAccrual) Start(ctx context.Context) error {

	return nil
}

func (a *HandlersAccrual) Stop(ctx context.Context) error {

	return nil
}
