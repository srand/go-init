package core

import (
	"fmt"
	"log"

	"github.com/srand/go-init/pkg/config"
	"github.com/srand/go-init/pkg/monitors"
	"github.com/srand/go-init/pkg/state"
	"github.com/srand/go-init/pkg/utils"
)

const (
	// Service running, pidfile exists
	ServiceStatusRunning = "running"

	// Service manually stopped.
	ServiceStatusStopped = "stopped"

	// Process running, but no pidfile exists yet
	ServiceStatusStarting = "starting"

	// Process has been sent SIGTERM but has not yet terminated.
	ServiceStatusStopping = "stopping"

	// Service is misconfigured and cannot be started
	ServiceStatusError = "error"
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
	CGroup        *ControlGroup
	Config        *config.ConfigService
	PidFile       *PidFile
	Preconditions []string

	Errors   state.State[int]
	Restarts state.State[int]
	Status   state.State[string]

	Actions    []state.Action
	Conditions []state.Condition
	Triggers   []state.Trigger

	actionChan chan ServiceAction
	process    *utils.Process
}

func NewService(registry state.ReferenceRegistry, svcConfig *config.ConfigService) (*Service, error) {
	service := &Service{
		Name:          svcConfig.Name,
		Command:       svcConfig.Command,
		Config:        svcConfig,
		PidFile:       NewPidFile(svcConfig),
		Preconditions: svcConfig.Conditions,
		actionChan:    make(chan ServiceAction),
	}

	service.Errors = state.NewState(
		fmt.Sprintf("services.%s.errors", service.Name),
		0,
	)

	service.Restarts = state.NewState(
		fmt.Sprintf("services.%s.restarts", service.Name),
		0,
	)

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

	runnable := state.NewCompositeCondition(fmt.Sprintf("services.%s.runnable", service.Name), false, true)
	for _, condName := range service.Config.Conditions {
		runnable.AddCondition(state.NewConditionRef(registry, condName))
	}
	service.Conditions = append(service.Conditions, runnable)

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

	stoppedOrError := state.NewCompositeCondition("stop|error", true, false)
	stoppedOrError.AddCondition(service.Conditions[0])
	stoppedOrError.AddCondition(service.Conditions[4])

	startingOrRunning := state.NewCompositeCondition("", true, false)
	startingOrRunning.AddCondition(service.Conditions[2])
	startingOrRunning.AddCondition(service.Conditions[3])

	startable := state.NewCompositeCondition(fmt.Sprintf("services.%s.startable", service.Name), false, false)
	startable.AddCondition(runnable)
	startable.AddCondition(stoppedOrError)
	service.Conditions = append(service.Conditions, startable)

	stoppable := state.NewCompositeCondition(fmt.Sprintf("services.%s.stoppable", service.Name), false, false)
	stoppable.AddCondition(state.NewNotCondition(runnable))
	stoppable.AddCondition(startingOrRunning)
	service.Conditions = append(service.Conditions, stoppable)

	service.Triggers = append(service.Triggers, state.NewActionTrigger(
		fmt.Sprintf("services.%s.trigger.start", service.Name),
		startable,
		service.Actions[0],
		nil,
	))
	service.Triggers = append(service.Triggers, state.NewActionTrigger(
		fmt.Sprintf("services.%s.trigger.stop", service.Name),
		stoppable,
		service.Actions[1],
		nil,
	))

	if svcConfig.CGroup != nil {
		var err error
		service.CGroup, err = NewDerivedControlGroup(registry, svcConfig.CGroup, service.Name)
		if err != nil {
			return nil, err
		}
	}

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
	s.process = utils.NewProcess(s.Command)
	if s.CGroup != nil {
		s.process.CGroup = s.CGroup.Name()
	}
	return s.process.Start()
}

func (s *Service) kill() {
	s.process.Terminate()
}

func (s *Service) pid() int {
	if s.process != nil {
		return s.process.Pid()
	}
	return 0
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
			if event.Pid != s.pid() {
				continue
			}

			s.process = nil
			if event.Status != 0 {
				s.setStatus(ServiceStatusError)
				s.Errors.Set(s.Errors.Get() + 1)
			} else {
				s.setStatus(ServiceStatusStopped)
			}

		case event := <-pidfileChan:
			if event.Name == s.PidFile.Path {
				switch s.Status.Get() {
				case ServiceStatusStarting:
					if pid := s.PidFile.Get(); pid == s.pid() {
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
				// log.Println("start:", s.Name, s.Status.Get())
				switch s.Status.Get() {
				case ServiceStatusStopped:
					fallthrough
				case ServiceStatusError:
					s.Restarts.Set(s.Restarts.Get() + 1)

					err := s.spawn()
					if err != nil {
						log.Println("Service failed to start:", s.Name, err.Error())
						s.setStatus(ServiceStatusError)
						break
					}
					s.setStatus(ServiceStatusStarting)

					if s.PidFile.Create {
						s.PidFile.Write(s.pid())
					}
				default:
					log.Println("Cannot start service in current state:", s.Name)
				}
			case ServiceActionStop:
				// log.Println("stop:", s.Name, s.Status.Get())
				switch s.Status.Get() {
				case ServiceStatusStarting:
				case ServiceStatusRunning:
					s.kill()
				}
			}
		}
	}
}
