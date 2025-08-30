package viewer

import (
	"fmt"
	"sync"
)

var lock = &sync.Mutex{}
var singleInstance *GinViewer

func GetViewerControllerInstance() *GinViewer {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		singleInstance = &GinViewer{}
	}

	return singleInstance
}

func SetViewerControllerInstance(userController *GinViewer) {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		fmt.Println("Creating single instance now.")
		singleInstance = userController
	}
}
