package adapter

import (
	"sync"

	"base/internal/models"
	"base/internal/service"
)

// Registry 子应用注册表，缓存应用配置
type Registry struct {
	mu   sync.RWMutex
	apps map[string]*models.App
}

var DefaultRegistry = &Registry{apps: make(map[string]*models.App)}

func (r *Registry) Register(app *models.App) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.apps[app.Code] = app
}

func (r *Registry) Get(code string) (*models.App, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	app, ok := r.apps[code]
	return app, ok
}

func (r *Registry) Reload() error {
	svc := service.AppService{}
	apps, err := svc.ListAllActive()
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.apps = make(map[string]*models.App)
	for _, app := range apps {
		a := app
		r.apps[a.Code] = &a
	}
	return nil
}
