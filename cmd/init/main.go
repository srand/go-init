package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/srand/go-init/pkg/config"
	"github.com/srand/go-init/pkg/monitors"
	"github.com/srand/go-init/pkg/services"
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
	services.SetPidFileDir(pidfileDir)

	configFile, err := config.ParseFile(configFilePath)
	if err != nil {
		panic(err)
	}

	registry, err := services.NewRegistry(configFile)
	if err != nil {
		panic(err)
	}

	for _, service := range registry.Services {
		fmt.Println(service.Name)
	}

	pidfileMonitor, err := monitors.NewFileMonitor(pidfileDir)
	if err != nil {
		panic(err)
	}
	go pidfileMonitor.Supervise()

	processMonitor := monitors.NewProcessMonitor()

	for _, svc := range registry.Services {
		go svc.Supervise(processMonitor, pidfileMonitor)
		go svc.Start()
	}

	processMonitor.Supervise()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

}
