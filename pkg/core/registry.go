package core

import (
	"fmt"
	"log"
	"sync"

	"github.com/srand/go-init/pkg/config"
	"github.com/srand/go-init/pkg/state"
)

type Registry struct {
	Services      map[string]*Service
	Tasks         map[string]*Task
	Actions       map[string]state.Action
	Conditions    map[string]state.Condition
	Triggers      map[string]state.Trigger
	observers     map[state.ReferenceObserver]struct{}
	observerMutex sync.Mutex
}

func (r *Registry) AddService(svc *Service) {
	r.Services[svc.Name] = svc
	r.PublishReference(svc.Name, svc)

	for _, action := range svc.Actions {
		r.Actions[action.Name()] = action
	}

	for _, cond := range svc.Conditions {
		r.AddCondition(cond)
	}

	for _, trig := range svc.Triggers {
		r.AddTrigger(trig)
	}
}

func (r *Registry) AddTask(task *Task) {
	r.Tasks[task.Name] = task
	r.PublishReference(task.Name, task)

	for _, action := range task.Actions {
		r.Actions[action.Name()] = action
	}

	for _, cond := range task.Conditions {
		r.AddCondition(cond)
	}

	for _, trig := range task.Triggers {
		r.AddTrigger(trig)
	}
}

func (r *Registry) FindAction(name string) state.Action {
	action, _ := r.Actions[name]
	return action
}

func (r *Registry) AddCondition(cond state.Condition) {
	log.Println("+", cond.Name())
	r.Conditions[cond.Name()] = cond
	r.PublishReference(cond.Name(), cond)
}

func (r *Registry) FindCondition(name string) state.Condition {
	cond, _ := r.Conditions[name]
	return cond
}

func (r *Registry) AddTrigger(trig state.Trigger) {
	log.Println("+", trig.Name())
	r.Triggers[trig.Name()] = trig
	r.PublishReference(trig.Name(), trig)
	trig.Eval()
}

func (r *Registry) SubscribeReference(name string, observer state.ReferenceObserver) {
	r.observerMutex.Lock()
	defer r.observerMutex.Unlock()
	r.observers[observer] = struct{}{}

	if obj, ok := r.Actions[name]; ok {
		observer.OnReferenceFound(name, obj)
		return
	}

	if obj, ok := r.Conditions[name]; ok {
		observer.OnReferenceFound(name, obj)
		return
	}

	if obj, ok := r.Services[name]; ok {
		observer.OnReferenceFound(name, obj)
		return
	}

	if obj, ok := r.Tasks[name]; ok {
		observer.OnReferenceFound(name, obj)
		return
	}
}

func (r *Registry) UnsubscribeReference(name string, observer state.ReferenceObserver) {
	r.observerMutex.Lock()
	defer r.observerMutex.Unlock()
	delete(r.observers, observer)
}

func (r *Registry) PublishReference(name string, obj any) {
	r.observerMutex.Lock()
	defer r.observerMutex.Unlock()

	for observer := range r.observers {
		observer.OnReferenceFound(name, obj)
	}
}

func (r *Registry) UnpublishReference(name string, obj any) {
	r.observerMutex.Lock()
	defer r.observerMutex.Unlock()

	for observer := range r.observers {
		observer.OnReferenceLost(name, obj)
	}
}

func NewRegistry(config *config.ConfigFile) (*Registry, error) {
	registry := &Registry{
		Actions:    map[string]state.Action{},
		Conditions: map[string]state.Condition{},
		Services:   map[string]*Service{},
		Tasks:      map[string]*Task{},
		Triggers:   map[string]state.Trigger{},
		observers:  map[state.ReferenceObserver]struct{}{},
	}

	for _, svcConfig := range config.Services {
		svc, err := NewService(registry, svcConfig)
		if err != nil {
			return nil, err
		}

		registry.AddService(svc)
	}

	for _, taskConfig := range config.Tasks {
		task, err := NewTask(registry, taskConfig)
		if err != nil {
			return nil, err
		}

		registry.AddTask(task)
	}

	// Create configured triggers
	for name, svc := range registry.Services {
		for _, trigConfig := range svc.Config.Triggers {
			cond := state.NewCompositeCondition("", false, false)

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
