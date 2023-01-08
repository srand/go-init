package core

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/srand/go-init/pkg/monitors"
	"github.com/srand/go-init/pkg/state"
)

type ModuleLoader struct {
	conditions map[string]*state.SimpleCond
	registry   state.ReferenceRegistry
}

func NewModuleLoader(registry state.ReferenceRegistry, monitor *monitors.ModuleMonitor) *ModuleLoader {
	loader := &ModuleLoader{
		conditions: map[string]*state.SimpleCond{},
		registry:   registry,
	}
	monitor.Subscribe(loader)
	loader.parseProcModules()
	return loader
}

func (l *ModuleLoader) addModule(module string) *state.SimpleCond {
	cond := state.NewCondition(fmt.Sprintf("modules.%s.loaded", module), nil)
	l.conditions[module] = cond
	l.registry.AddReference(cond)
	return cond
}

func (l *ModuleLoader) parseProcModules() error {
	file, err := os.Open("/proc/modules")
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineno := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineno++
		module, _, _ := strings.Cut(line, " ")
		l.OnModuleAdded(module)
	}

	return nil
}

func (l *ModuleLoader) OnModuleAdded(module string) {
	var cond *state.SimpleCond
	var ok bool

	if cond, ok = l.conditions[module]; !ok {
		cond = l.addModule(module)
	}

	cond.Set(true)
}

func (l *ModuleLoader) OnModuleRemoved(module string) {
	var cond *state.SimpleCond
	var ok bool

	if cond, ok = l.conditions[module]; !ok {
		cond = l.addModule(module)
	}

	cond.Set(false)
}
