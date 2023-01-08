package core

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/srand/go-init/pkg/config"
	"github.com/srand/go-init/pkg/monitors"
	"github.com/srand/go-init/pkg/state"
)

type ModuleLoader struct {
	conditions map[string]*state.SimpleCond
	registry   state.ReferenceRegistry
}

func NewModuleLoader(registry state.ReferenceRegistry, monitor *monitors.ModuleMonitor, config *config.ConfigModules) *ModuleLoader {
	loader := &ModuleLoader{
		conditions: map[string]*state.SimpleCond{},
		registry:   registry,
	}
	monitor.Subscribe(loader)
	loader.parseProcModules()
	fmt.Println(config)
	loader.loadModules(config)
	loader.loadIncludes(config)
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

func (l *ModuleLoader) parseConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineno := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineno++

		if strings.HasPrefix(line, "#") {
			continue
		}

		module, _, _ := strings.Cut(line, " ")

		if err := l.loadModule(module); err != nil {
			log.Println("E", "modules."+module)
		}
	}

	return nil
}

func (l *ModuleLoader) loadModules(config *config.ConfigModules) error {
	for _, module := range config.Modules {
		err := l.loadModule(module.Name, module.Parameters...)
		if err != nil {
			log.Println("E", "modules."+module.Name)
		}
	}
	return nil
}

func (l *ModuleLoader) loadIncludes(config *config.ConfigModules) error {
	for _, incpath := range config.IncludePaths {
		files, err := filepath.Glob(incpath)
		if err != nil {
			continue
		}

		for _, file := range files {
			l.parseConfig(file)
		}
	}
	return nil
}

func (l *ModuleLoader) loadModule(module string, parameters ...string) error {
	params := []string{module}
	params = append(params, parameters...)
	cmd := exec.Command("modprobe", params...)
	return cmd.Run()
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
