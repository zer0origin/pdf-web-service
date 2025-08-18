package user

import (
	"fmt"
	"sync"
)

var lock = &sync.Mutex{}
var singleInstance *GinUser

func GetControllerInstance() *GinUser {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		singleInstance = &GinUser{}
	}

	return singleInstance
}

func SetControllerInstance(userController *GinUser) {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		fmt.Println("Creating single instance now.")
		singleInstance = userController
	}
}
