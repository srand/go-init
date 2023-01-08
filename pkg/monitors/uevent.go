package monitors

import (
	"github.com/srand/go-init/pkg/state"
	"github.com/srand/go-init/pkg/utils"
)

type UEventMonitor struct {
	conn      utils.UEventConn
	observers *state.SimpleObservable[*UEventMonitor, *utils.UEvent]
}

func NewUEventMonitor() (*UEventMonitor, error) {
	monitor := &UEventMonitor{
		observers: state.NewObservable[*UEventMonitor, *utils.UEvent](),
	}

	if err := monitor.conn.Dial(); err != nil {
		return nil, err
	}

	return monitor, nil
}

func (m *UEventMonitor) Subscribe(observer state.Observer[*UEventMonitor, *utils.UEvent]) {
	m.observers.Subscribe(observer)
}

func (m *UEventMonitor) Unsubscribe(observer state.Observer[*UEventMonitor, *utils.UEvent]) {
	m.observers.Unsubscribe(observer)
}

func (m *UEventMonitor) Close() {
	m.conn.Close()
}

func (m *UEventMonitor) Supervise() {
	for {
		event, err := m.conn.ReadEvent()
		if err != nil {
			m.conn.Close()
			return
		}

		if event.Name == "libudev" {
			continue
		}

		m.observers.Publish(m, event)
	}
}
