package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/srand/go-init/pkg/config"
	"github.com/srand/go-init/pkg/state"
)

const (
	ControlGroupPath = "/sys/fs/cgroup"
)

type ControlGroup struct {
	name       string
	Conditions []state.Condition
	Config     *config.ConfigControlGroup
}

func NewControlGroup(config *config.ConfigControlGroup) (*ControlGroup, error) {
	cg := &ControlGroup{
		name:   config.Name,
		Config: config,
	}

	if err := cg.create(); err != nil {
		return nil, err
	}

	hasError := false

	for key, value := range config.Config {
		if err := cg.set(key, value); err != nil {
			hasError = true
		}
	}

	cg.Conditions = append(cg.Conditions, state.NewCondition(
		fmt.Sprintf("cgroups.%s.configured", config.Name),
		func(update func(bool)) {
			update(!hasError)
		},
	))

	return cg, nil
}

func NewControlGroupRef(name string) (*ControlGroup, error) {
	cg := &ControlGroup{name: name}
	if !cg.exists() {
		return nil, errors.New("no such control group")
	}
	return cg, nil
}

func NewDerivedControlGroup(registry state.ReferenceRegistry, cgConfig *config.ConfigControlGroup, name string) (*ControlGroup, error) {
	cgref := registry.FindReference(cgConfig.Name)
	if cgref == nil {
		return nil, fmt.Errorf("No such control group: %s", cgConfig.Name)
	}

	cg, ok := cgref.(*ControlGroup)
	if !ok {
		return nil, fmt.Errorf("No such control group: %s", cgConfig.Name)
	}

	newcg, err := cg.Derive(name, cgConfig.Config)
	if err != nil {
		return nil, fmt.Errorf("Error creating control group: %v", err)
	}

	return newcg, nil
}

func RootControlGroup() *ControlGroup {
	return &ControlGroup{name: ""}
}

func (cg *ControlGroup) Name() string {
	return cg.name
}

func (cg *ControlGroup) AddPid(pid int) error {
	return cg.set("cgroup.procs", fmt.Sprintf("%d\n", pid))
}

func (cg *ControlGroup) Dispose() error {
	return cg.remove()
}

func (cg *ControlGroup) Controllers() ([]string, error) {
	data, err := os.ReadFile(filepath.Join(cg.path(), "cgroup.controllers"))
	if err != nil {
		return []string{}, nil
	}

	return strings.Split(string(data), " "), nil
}

func (cg *ControlGroup) ExportControllers(controllers []string) error {
	for _, name := range controllers {
		err := cg.set("cgroup.subtree_control", "+"+name+"\n")
		if err != nil {
			if errors.Is(err, syscall.ENOTSUP) {
				continue
			}
		}
	}
	return nil
}

func (cg *ControlGroup) Derive(name string, config map[string]string) (*ControlGroup, error) {
	cgd := &ControlGroup{
		name: cg.name + "." + strings.ReplaceAll(name, ".", "_"),
	}

	if err := cgd.create(); err != nil {
		return nil, err
	}

	for key, value := range config {
		if err := cgd.set(key, value); err != nil {
			cgd.remove()
			return nil, err
		}
	}

	return cgd, nil
}

func (cg *ControlGroup) path() string {
	return filepath.Join(ControlGroupPath, strings.ReplaceAll(cg.name, ".", "/"))
}

func (cg *ControlGroup) keypath(key string) string {
	return filepath.Join(cg.path(), key)
}

func (cg *ControlGroup) exists() bool {
	if _, err := os.Stat(cg.path()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (cg *ControlGroup) create() error {
	return os.MkdirAll(cg.path(), 0777)
}

func (cg *ControlGroup) remove() error {
	return syscall.Rmdir(cg.path())
}

func (cg *ControlGroup) set(key, value string) error {
	return os.WriteFile(cg.keypath(key), []byte(value+"\n"), 0755)
}
