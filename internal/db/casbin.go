package db

import (
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/wuyoushe/hyper-go/library/conf/paladin"
	"github.com/wuyoushe/hyper-go/library/database"
	"github.com/wuyoushe/hyper-go/library/log"
	"time"

	"github.com/wuyoushe/hyper-go/tool\hyper/felton_blog/internal/config"
	"github.com/wuyoushe/hyper-go/tool\hyper/felton_blog/internal/service"
)

// NewCasbin after Service, and SetEnforcer for service
func NewCasbin(svc service.Service, cfg *config.Config) (ef *casbin.SyncedEnforcer, err error) {
	key := "casbin.toml"
	if err = paladin.Get(key).UnmarshalTOML(cfg); err != nil {
		return
	}
	ef, err = database.NewCasbinConn(svc, cfg.ConfigPath, cfg.Casbin)
	if err != nil {
		return
	}
	if err = svc.SetEnforcer(ef); err != nil {
		return
	}
	go WatchCasbinModel(ef, cfg.Casbin)
	go WatchCasbinConfig(ef, cfg.Casbin, key)
	return
}

func WatchCasbinModel(e *casbin.SyncedEnforcer, c *database.CasbinConfig) {
	for range paladin.WatchEvent(context.Background(), c.Model) {
		if err := e.LoadModel(); err != nil {
			log.Printf("e.LoadModel error(%v)", err)
		}
	}
}

func WatchCasbinConfig(e *casbin.SyncedEnforcer, c *database.CasbinConfig, key string) {
	for event := range paladin.WatchEvent(context.Background(), key) {
		autoLoad := c.AutoLoad
		autoLoadInternal := c.AutoLoadInternal
		s := &paladin.TOML{}
		_ = s.Set(event.Value)
		if err := s.Get("Casbin").UnmarshalTOML(c); err != nil {
			continue
		}
		if c.AutoLoad != autoLoad {
			if c.AutoLoad {
				_ = e.InitWithModelAndAdapter(e.GetModel(), e.GetAdapter())
				e.StartAutoLoadPolicy(time.Duration(c.AutoLoadInternal) * time.Second)
			} else {
				e.StopAutoLoadPolicy()
			}
		}
		if c.AutoLoadInternal != autoLoadInternal {
			e.StopAutoLoadPolicy()
			_ = e.InitWithModelAndAdapter(e.GetModel(), e.GetAdapter())
			e.StartAutoLoadPolicy(time.Duration(c.AutoLoadInternal) * time.Second)
		}
		if !c.Enable {
			e.StopAutoLoadPolicy()
		}
	}
}
