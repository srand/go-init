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
	// Task running, pidfile exists
	TaskStatusRunning = "running"

	// Task manually stopped.
	TaskStatusPending = "pending"

	// Task has completed
	TaskStatusCompleted = "completed"

	// Process has been sent SIGTERM but has not yet terminated.
	TaskStatusStopping = "stopping"

	// Task is misconfigured and cannot be started
	TaskStatusError = "error"
)

const (
	TaskActionStart = iota
	TaskActionStop
)

type TaskAction int
type TaskStatus string

type Task struct {
	Name    string
	CGroup  *ControlGroup
	Command []string
	Config  *config.ConfigTask
	Status  state.State[string]

	Actions    []state.Action
	Conditions []state.Condition
	Triggers   []state.Trigger

	actionChan chan TaskAction
	process    *utils.Process
}

func NewTask(registry state.ReferenceRegistry, svcConfig *config.ConfigTask) (*Task, error) {
	task := &Task{
		Name:       svcConfig.Name,
		Command:    svcConfig.Command,
		Config:     svcConfig,
		actionChan: make(chan TaskAction),
	}

	task.Status = state.NewState(
		fmt.Sprintf("tasks.%s.state", task.Name),
		TaskStatusPending,
	)

	task.Conditions = append(task.Conditions, state.NewStateCondition(
		fmt.Sprintf("tasks.%s.state.pending", task.Name),
		task.Status,
		TaskStatusPending,
	))

	task.Conditions = append(task.Conditions, state.NewStateCondition(
		fmt.Sprintf("tasks.%s.state.stopping", task.Name),
		task.Status,
		TaskStatusStopping,
	))

	task.Conditions = append(task.Conditions, state.NewStateCondition(
		fmt.Sprintf("tasks.%s.state.running", task.Name),
		task.Status,
		TaskStatusRunning,
	))

	task.Conditions = append(task.Conditions, state.NewStateCondition(
		fmt.Sprintf("tasks.%s.state.completed", task.Name),
		task.Status,
		TaskStatusCompleted,
	))

	task.Conditions = append(task.Conditions, state.NewStateCondition(
		fmt.Sprintf("tasks.%s.state.error", task.Name),
		task.Status,
		TaskStatusError,
	))

	runnable := state.NewCompositeCondition(fmt.Sprintf("tasks.%s.runnable", task.Name), false, true)
	for _, condName := range task.Config.Conditions {
		runnable.AddCondition(state.NewConditionRef(registry, condName))
	}
	task.Conditions = append(task.Conditions, runnable)

	task.Actions = append(task.Actions, state.NewAction(
		fmt.Sprintf("tasks.%s.action.start", task.Name),
		func() error {
			go task.Start()
			return nil
		},
	))

	task.Actions = append(task.Actions, state.NewAction(
		fmt.Sprintf("tasks.%s.action.stop", task.Name),
		func() error {
			go task.Stop()
			return nil
		},
	))

	runCondition := state.NewCompositeCondition("", false, false)
	runCondition.AddCondition(runnable)
	runCondition.AddCondition(task.Conditions[0]) // Status must be pending

	task.Triggers = append(task.Triggers, state.NewActionTrigger(
		fmt.Sprintf("tasks.%s.trigger.runnable", task.Name),
		runCondition,
		task.Actions[0],
		nil,
	))

	if svcConfig.CGroup != nil {
		var err error
		task.CGroup, err = NewDerivedControlGroup(registry, svcConfig.CGroup, task.Name)
		if err != nil {
			return nil, err
		}
	}

	return task, nil
}

func (s *Task) Start() {
	s.actionChan <- TaskActionStart
}

func (s *Task) Stop() {
	s.actionChan <- TaskActionStop
}

func (s *Task) setStatus(status TaskStatus) {
	s.Status.Set(string(status))
}

func (s *Task) spawn() error {
	s.process = utils.NewProcess(s.Command)
	if s.CGroup != nil {
		s.process.CGroup = s.CGroup.Name()
	}
	return s.process.Start()
}

func (s *Task) kill() {
	s.process.Terminate()
}

func (s *Task) pid() int {
	if s.process != nil {
		return s.process.Pid()
	}
	return 0
}
func (s *Task) FindAction(name string) state.Action {
	for _, action := range s.Actions {
		if action.Name() == fmt.Sprintf("tasks.%s.action.%s", s.Name, name) {
			return action
		}
	}
	return nil
}

func (s *Task) AddTrigger(trig state.Trigger) {
	s.Triggers = append(s.Triggers, trig)
}

func (s *Task) Supervise(procMonitor *monitors.ProcessMonitor) {
	procChan := procMonitor.Subscribe()

	for {
		select {
		case event := <-procChan:
			if event.Pid != s.pid() {
				continue
			}

			s.process = nil
			if event.Status != 0 {
				s.setStatus(TaskStatusError)
			} else {
				s.setStatus(TaskStatusCompleted)
			}

		case cmd := <-s.actionChan:
			// log.Println("Task status change requested:", s.Name, s.Status, status)
			switch cmd {
			case TaskActionStart:
				// Requested to start Task, act according to current status
				switch s.Status.Get() {
				case TaskStatusPending:
					err := s.spawn()
					if err != nil {
						log.Println("Task failed to start:", s.Name, err.Error())
						s.setStatus(TaskStatusError)
						break
					}
					s.setStatus(TaskStatusRunning)
				default:
					log.Println("Cannot start task in current state:", s.Name)
				}
			case TaskActionStop:
				switch s.Status.Get() {
				case TaskStatusRunning:
					s.setStatus(TaskStatusStopping)
					s.kill()
				}
			}
		}
	}
}
