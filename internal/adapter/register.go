package adapter

import (
	"github.com/user/subscriptions-monitor/internal/adapter/kimi"
	"github.com/user/subscriptions-monitor/internal/adapter/minimax"
	"github.com/user/subscriptions-monitor/internal/adapter/zenmux"
	"github.com/user/subscriptions-monitor/internal/provider"
)

func RegisterAll(r *provider.Registry) {
	r.Register(zenmux.New())
	r.Register(kimi.New())
	r.Register(minimax.New())
}
