package services

import (
	"fmt"
	"log"

	"github.com/srand/go-init/pkg/config"
	"github.com/srand/go-init/pkg/state"
)

type Registry struct {
	Services   map[string]*Service
	Actions    map[string]state.Action
	Conditions map[string]state.Condition
	Triggers   map[string]state.Trigger
}

func (r *Registry) AddService(svc *Service) {
	r.Services[svc.Name] = svc

	for _, action := range svc.Actions {
		r.Actions[action.Name()] = action
	}

	for _, cond := range svc.Conditions {
		r.AddCondition(cond)
	}

	for _, trig := range svc.Triggers {
		r.Triggers[trig.Name()] = trig
	}
}

func (r *Registry) FindAction(name string) state.Action {
	action, _ := r.Actions[name]
	return action
}

func (r *Registry) AddCondition(cond state.Condition) {
	log.Println("New condition:", cond.Name())
	r.Conditions[cond.Name()] = cond
}

func (r *Registry) FindCondition(name string) state.Condition {
	cond, _ := r.Conditions[name]
	return cond
}

func (r *Registry) AddTrigger(trig state.Trigger) {
	log.Println("New trigger:", trig.Name())
	r.Triggers[trig.Name()] = trig
	trig.Eval()
}

func NewRegistry(config *config.ConfigFile) (*Registry, error) {
	registry := &Registry{
		Services:   map[string]*Service{},
		Actions:    map[string]state.Action{},
		Conditions: map[string]state.Condition{},
		Triggers:   map[string]state.Trigger{},
	}

	for _, svcConfig := range config.Services {
		svc, err := NewService(svcConfig)
		if err != nil {
			return nil, err
		}

		registry.AddService(svc)
	}

	// Create default triggers if none is configured
	for name, svc := range registry.Services {
		if len(svc.Config.Triggers) > 0 {
			// Manual triggers present, skip default behavior
			continue
		}

		startAction := svc.FindAction("start")
		if startAction == nil {
			panic("No action 'start' found in service: " + name)
		}

		stopAction := svc.FindAction("stop")
		if stopAction == nil {
			panic("No action 'stop' found in service: " + name)
		}

		preCond := state.NewCompositeCondition("")
		for _, precondName := range svc.Preconditions {
			cond := registry.FindCondition(precondName)
			if cond == nil {
				panic("Precondition not found: " + precondName)
			}
			preCond.AddCondition(cond)
		}

		name = fmt.Sprintf("services.%s.trigger.preconditions", name)
		trig := state.NewActionTrigger(name, preCond, startAction, stopAction)

		svc.AddTrigger(trig)
		registry.AddTrigger(trig)
	}

	// Create configured triggers
	for name, svc := range registry.Services {
		for _, trigConfig := range svc.Config.Triggers {
			cond := state.NewCompositeCondition("")

			for _, trigCondName := range trigConfig.Conditions {
				trigCond := registry.FindCondition(trigCondName)
				if trigCond == nil {
					panic("Trigger precondition not found: " + trigCondName)
				}
				cond.AddCondition(trigCond)
			}

			// FIXME: Composite action
			actionName := trigConfig.Actions[0]
			action := registry.FindAction(actionName)
			if action == nil {
				panic("Trigger action not found: " + actionName)
			}

			trigName := fmt.Sprintf("services.%s.trigger.%s", name, trigConfig.Name)
			trig := state.NewActionTrigger(trigName, cond, action, nil)
			registry.AddTrigger(trig)
		}
	}

	return registry, nil
}
