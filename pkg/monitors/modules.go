package monitors

import (
	"strings"

	"github.com/srand/go-init/pkg/utils"
)

type ModuleMonitor struct {
	observers map[ModuleObserver]struct{}
}

type ModuleObserver interface {
	OnModuleAdded(string)
	OnModuleRemoved(string)
}

func NewModuleMonitor(um *UEventMonitor) *ModuleMonitor {
	monitor := &ModuleMonitor{
		observers: map[ModuleObserver]struct{}{},
	}

	um.Subscribe(monitor)

	return monitor
}

func (m *ModuleMonitor) Subscribe(o ModuleObserver) {
	m.observers[o] = struct{}{}
}

func (m *ModuleMonitor) Unsubscribe(o ModuleObserver) {
	delete(m.observers, o)
}

func (m *ModuleMonitor) OnChange(obj *UEventMonitor, event *utils.UEvent) {
	// fmt.Println(event)

	if subsys, ok := event.Env["SUBSYSTEM"]; !ok || subsys != "module" {
		return
	}

	devpath, ok := event.Env["DEVPATH"]
	if !ok {
		return
	}

	action, ok := event.Env["ACTION"]
	if !ok {
		return
	}

	if !strings.HasPrefix(devpath, "/module/") {
		return
	}

	module := devpath[8:]

	switch action {
	case "add":
		for obs := range m.observers {
			obs.OnModuleAdded(module)
		}
	case "remove":
		for obs := range m.observers {
			obs.OnModuleRemoved(module)
		}
	}
}
