package events

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Watcher struct {
	Client     client.Client
	projectMap map[string]interface{}
}

func (w *Watcher) Watch() {
	// This is terribly inefficient. But, we're going to just run with it for now.

}
