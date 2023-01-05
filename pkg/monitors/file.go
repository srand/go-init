package monitors

import (
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/srand/go-init/pkg/utils"
)

type FileEvent fsnotify.Event

type FileMonitor struct {
	Path   string
	broker *utils.Broker[FileEvent]
}

func NewFileMonitor(path string) (*FileMonitor, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	return &FileMonitor{
		Path:   path,
		broker: utils.NewBroker[FileEvent](),
	}, nil
}

func (f *FileMonitor) Subscribe() chan FileEvent {
	return f.broker.Subscribe()
}

func (f *FileMonitor) Unsubscribe(channel chan FileEvent) {
	f.broker.Unsubscribe(channel)
}

func (f *FileMonitor) Supervise() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	err = watcher.Add(f.Path)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			f.broker.Publish(FileEvent(event))
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Fatal("error:", err)
		}
	}

}
