package monitors

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/srand/go-init/pkg/utils"
)

type ProcessTerminationEvent struct {
	Pid    int
	Status int
}

type ProcessMonitor struct {
	broker  *utils.Broker[ProcessTerminationEvent]
	sigchan chan os.Signal
}

func NewProcessMonitor() *ProcessMonitor {
	return &ProcessMonitor{
		utils.NewBroker[ProcessTerminationEvent](),
		make(chan os.Signal, 100),
	}
}

func (m *ProcessMonitor) Subscribe() chan ProcessTerminationEvent {
	return m.broker.Subscribe()
}

func (m *ProcessMonitor) Unsubscribe(channel chan ProcessTerminationEvent) {
	m.broker.Unsubscribe(channel)
}

func (m *ProcessMonitor) Supervise() {
	signal.Notify(m.sigchan, syscall.SIGCHLD)

	for {
		<-m.sigchan

		var status syscall.WaitStatus

		wpid, err := syscall.Wait4(-1, &status, 0, nil)
		if err != nil {
			continue
		}

		// log.Printf("Process terminated: %d (%d)\n", wpid, status)

		m.broker.Publish(ProcessTerminationEvent{
			Pid:    wpid,
			Status: int(status),
		})
	}
}
