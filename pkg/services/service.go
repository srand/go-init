package services

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"

	"github.com/srand/go-init/pkg/config"
	"github.com/srand/go-init/pkg/monitors"
	"github.com/srand/go-init/pkg/state"
)

const (
	// Service running, pidfile exists
	ServiceStatusRunning = "Running"

	// Service manually stopped.
	ServiceStatusStopped = "Stopped"

	// Process running, but no pidfile exists yet
	ServiceStatusStarting = "Starting"

	// Process has been sent SIGTERM but has not yet terminated.
	ServiceStatusStopping = "Stopping"

	// Service is misconfigured and cannot be started
	ServiceStatusError = "Error"
)

const (
	ServiceActionStart = iota
	ServiceActionStop
)

type ServiceAction int
type ServiceStatus string

type Service struct {
	Name          string
	Command       []string
	Config        *config.ConfigService
	Pid           int
	PidFile       *PidFile
	Preconditions []string
	Status        state.State[string]

	Actions    []state.Action
	Conditions []state.Condition
	Triggers   []state.Trigger

	actionChan chan ServiceAction
	command    *exec.Cmd
}

func NewService(svcConfig *config.ConfigService) (*Service, error) {
	service := &Service{
		Name:          svcConfig.Name,
		Command:       svcConfig.Command,
		Config:        svcConfig,
		Pid:           0,
		PidFile:       NewPidFile(svcConfig),
		Preconditions: svcConfig.Conditions,
		actionChan:    make(chan ServiceAction),
	}

	service.Status = state.NewState(
		fmt.Sprintf("services.%s.state", service.Name),
		ServiceStatusStopped,
	)

	service.Conditions = append(service.Conditions, state.NewStateCondition(
		fmt.Sprintf("services.%s.state.stopped", service.Name),
		service.Status,
		ServiceStatusStopped,
	))

	service.Conditions = append(service.Conditions, state.NewStateCondition(
		fmt.Sprintf("services.%s.state.stopping", service.Name),
		service.Status,
		ServiceStatusStopping,
	))

	service.Conditions = append(service.Conditions, state.NewStateCondition(
		fmt.Sprintf("services.%s.state.starting", service.Name),
		service.Status,
		ServiceStatusStarting,
	))

	service.Conditions = append(service.Conditions, state.NewStateCondition(
		fmt.Sprintf("services.%s.state.running", service.Name),
		service.Status,
		ServiceStatusRunning,
	))

	service.Conditions = append(service.Conditions, state.NewStateCondition(
		fmt.Sprintf("services.%s.state.error", service.Name),
		service.Status,
		ServiceStatusError,
	))

	service.Actions = append(service.Actions, state.NewAction(
		fmt.Sprintf("services.%s.action.start", service.Name),
		func() error {
			go service.Start()
			return nil
		},
	))

	service.Actions = append(service.Actions, state.NewAction(
		fmt.Sprintf("services.%s.action.stop", service.Name),
		func() error {
			go service.Stop()
			return nil
		},
	))

	return service, nil
}

func (s *Service) Start() {
	s.actionChan <- ServiceActionStart
}

func (s *Service) Stop() {
	s.actionChan <- ServiceActionStop
}

func (s *Service) setStatus(status ServiceStatus) {
	s.Status.Set(string(status))
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

func (s *Service) FindAction(name string) state.Action {
	for _, action := range s.Actions {
		if action.Name() == fmt.Sprintf("services.%s.action.%s", s.Name, name) {
			return action
		}
	}
	return nil
}

func (s *Service) AddTrigger(trig state.Trigger) {
	s.Triggers = append(s.Triggers, trig)
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
			s.setStatus(ServiceStatusStopped)
			if event.Status != 0 {
			} else {
			}

		case event := <-pidfileChan:
			if event.Name == s.PidFile.Path {
				switch s.Status.Get() {
				case ServiceStatusStarting:
					if pid := s.PidFile.Get(); pid == s.Pid {
						s.setStatus(ServiceStatusRunning)
					}
				default:
				}
			}

		case cmd := <-s.actionChan:
			// log.Println("Service status change requested:", s.Name, s.Status, status)
			switch cmd {
			case ServiceActionStart:
				// Requested to start service, act according to current status
				switch s.Status.Get() {
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
					log.Println("Cannot start service in current state:", s.Name)
				}
			case ServiceActionStop:
				switch s.Status.Get() {
				case ServiceStatusStarting:
				case ServiceStatusRunning:
					s.kill()
				}
			}
		}
	}
}
