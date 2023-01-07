package core

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/srand/go-init/pkg/config"
	"github.com/srand/go-init/pkg/utils"
)

func SetSysctlParam(key, value string) bool {
	procroot := "/proc/sys"
	keypath := strings.ReplaceAll(key, ".", "/")
	path := filepath.Join(procroot, keypath)
	if err := os.WriteFile(path, []byte(value), 0755); err != nil {
		log.Println("E sysctl."+key, value)
		return false
	} else {
		log.Println("= sysctl."+key, value)
	}
	return true
}

func ReadSysctlProperties(filename string) (utils.Properties, error) {
	return utils.ReadPropertiesFile(filename)
}

func ApplySysctl(conf *config.ConfigSysctl) bool {
	for _, incpath := range conf.IncludePaths {
		files, err := filepath.Glob(incpath)
		if err != nil {
			continue
		}

		for _, file := range files {
			props, err := ReadSysctlProperties(file)
			if err != nil {
				continue
			}

			for key, value := range props {
				SetSysctlParam(key, value)
			}
		}
	}

	for _, param := range conf.Parameters {
		SetSysctlParam(param.Key, param.Value)
	}

	return true
}
