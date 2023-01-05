package services

import (
	"log"
	"os/exec"
	"syscall"

	"github.com/srand/go-init/pkg/config"
	"github.com/srand/go-init/pkg/monitors"
	"github.com/srand/go-init/pkg/utils"
)

const (
	// Newly created service, never started
	ServiceStatusCreated = "Created"

	// Service conditions have not been satisfied, service not running
	ServiceStatusBlocked = "Blocked"

	// Service running, pidfile exists
	ServiceStatusRunning = "Running"

	// Service manually stopped.
	ServiceStatusStopped = "Stopped"

	// Service has terminated with exit status != 0, may be restarted
	ServiceStatusCrashed = "Crashed"

	// Service will be restarted, not running
	ServiceStatusRestart = "Restart"

	// Process running, but no pidfile exists yet
	ServiceStatusStarting = "Starting"

	// Service is misconfigured and cannot be started
	ServiceStatusError = "Error"
)

type ServiceStatus string

type Service struct {
	Name         string
	Command      []string
	Pid          int
	PidFile      *PidFile
	Status       ServiceStatus
	StatusBroker *utils.Broker[ServiceStatus]

	statusChan chan ServiceStatus
	command    *exec.Cmd
}

func NewService(svcConfig *config.ConfigService) (*Service, error) {
	return &Service{
		Name:         svcConfig.Name,
		Command:      svcConfig.Command,
		Pid:          0,
		PidFile:      NewPidFile(svcConfig),
		Status:       ServiceStatusCreated,
		StatusBroker: utils.NewBroker[ServiceStatus](),
		statusChan:   make(chan ServiceStatus),
	}, nil
}

func (s *Service) Start() {
	s.statusChan <- ServiceStatusStarting
}

func (s *Service) Stop() {
	s.statusChan <- ServiceStatusStopped
}

func (s *Service) setStatus(status ServiceStatus) {
	log.Println("Service status change:", s.Name, status)

	s.Status = status

	// Notify subscribers about new service status
	s.StatusBroker.Publish(s.Status)
}

func (s *Service) spawn() error {
	cmd := exec.Command(s.Command[0], s.Command[1:]...)
	err := cmd.Start()
	if err != nil {
		return err
	}

	s.Pid = cmd.Process.Pid
	return nil
}

func (s *Service) kill() {
	syscall.Kill(s.Pid, syscall.SIGTERM)
}

func (s *Service) Supervise(procMonitor *monitors.ProcessMonitor, pidfileMonitor *monitors.FileMonitor) {
	procChan := procMonitor.Subscribe()
	pidfileChan := pidfileMonitor.Subscribe()

	for {
		select {
		case event := <-procChan:
			if event.Pid != s.Pid {
				continue
			}

			s.Pid = 0
			if event.Status != 0 {
				s.setStatus(ServiceStatusCrashed)
			} else {
				s.setStatus(ServiceStatusRestart)
				go s.Start()
			}

		case event := <-pidfileChan:
			if event.Name == s.PidFile.Path {
				switch s.Status {
				case ServiceStatusStarting:
					if pid := s.PidFile.Get(); pid == s.Pid {
						s.setStatus(ServiceStatusRunning)
					}
				default:
				}
			}

		case status := <-s.statusChan:
			// log.Println("Service status change requested:", s.Name, s.Status, status)
			switch status {
			case ServiceStatusStarting:
				// Requested to start service, act according to current status
				switch s.Status {
				case ServiceStatusCreated:
					fallthrough
				case ServiceStatusRestart:
					fallthrough
				case ServiceStatusStopped:
					err := s.spawn()
					if err != nil {
						log.Println("Service failed to start:", s.Name, err.Error())
						s.setStatus(ServiceStatusError)
						break
					}
					s.setStatus(ServiceStatusStarting)

					if s.PidFile.Create {
						s.PidFile.Write(s.Pid)
					}
				default:
					log.Println("Cannot start service in current state:", s.Name, status)
				}
			case ServiceStatusStopped:
				switch s.Status {
				case ServiceStatusStarting:
				case ServiceStatusRunning:
					s.kill()
				}
			}
		}
	}
}
