package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/srand/go-init/pkg/config"
)

type PidFile struct {
	Path   string
	Create bool
}

var pidFileDir string

func SetPidFileDir(dir string) {
	pidFileDir = dir
}

func NewPidFile(svcConfig *config.ConfigService) *PidFile {
	var path string
	var create bool

	if svcConfig.PidFile != nil {
		path = svcConfig.PidFile.Path
		create = svcConfig.PidFile.Create
	} else {
		// By default, assume the service is creating a pidfile itself
		create = false
	}
	if path == "" {
		path = filepath.Join(pidFileDir, svcConfig.Name+".pid")
	}

	return &PidFile{
		Path:   path,
		Create: create,
	}
}

func (f *PidFile) Get() int {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return 0
	}

	dataStr := strings.TrimSpace(string(data))

	pid, err := strconv.Atoi(dataStr)
	if err != nil {
		return 0
	}

	return pid
}

func (f *PidFile) Write(pid int) error {
	return os.WriteFile(f.Path, []byte(fmt.Sprint(pid)), fs.FileMode(0755))
}
