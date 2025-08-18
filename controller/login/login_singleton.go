package login

import (
	"fmt"
	"sync"
)

var lock = &sync.Mutex{}
var singleInstance *GinLogin

func GetControllerInstance() *GinLogin {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		singleInstance = &GinLogin{}
	}

	return singleInstance
}

func SetControllerInstance(userController *GinLogin) {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		fmt.Println("Creating single instance now.")
		singleInstance = userController
	}
}
