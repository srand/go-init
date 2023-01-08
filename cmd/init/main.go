package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/srand/go-init/pkg/config"
	"github.com/srand/go-init/pkg/core"
	"github.com/srand/go-init/pkg/monitors"
)

var rootCmd = cobra.Command{
	Use:   "init",
	Short: "Init system",
	Run:   run,
}

func init() {
	rootCmd.Flags().String("config", "/etc/init.yaml", "Configuration file")
	rootCmd.Flags().String("pidfile", "/var/run", "Pidfile directory")
}

func run(cmd *cobra.Command, args []string) {
	configFilePath, err := cmd.Flags().GetString("config")
	if err != nil {
		panic(err)
	}

	pidfileDir, err := cmd.Flags().GetString("pidfile")
	if err != nil {
		panic(err)
	}
	core.SetPidFileDir(pidfileDir)

	configFile, err := config.ParseFile(configFilePath)
	if err != nil {
		panic(err)
	}

	if configFile.Sysctl != nil && !core.ApplySysctl(configFile.Sysctl) {
		panic("failed to set system parameters")
	}

	rootCg := core.RootControlGroup()
	cgControllers, err := rootCg.Controllers()
	if err != nil {
		panic(err)
	}

	err = rootCg.ExportControllers(cgControllers)
	if err != nil {
		log.Println("warning: could not export process controllers", err.Error())
	}

	registry, err := core.NewRegistry(configFile)
	if err != nil {
		panic(err)
	}

	pidfileMonitor, err := monitors.NewFileMonitor(pidfileDir)
	if err != nil {
		panic(err)
	}
	go pidfileMonitor.Supervise()

	processMonitor := monitors.NewProcessMonitor()

	for _, svc := range registry.Services {
		go svc.Supervise(processMonitor, pidfileMonitor)
	}

	for _, task := range registry.Tasks {
		go task.Supervise(processMonitor)
	}

	processMonitor.Supervise()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

}
