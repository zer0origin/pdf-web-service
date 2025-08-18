package controller

import (
	"fmt"
	"sync"
)

var lock = &sync.Mutex{}
var singleInstance *GinMiddleware

func GetMiddlewareInstance() *GinMiddleware {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		singleInstance = &GinMiddleware{}
	}

	return singleInstance
}

func SetMiddlewareInstance(userController *GinMiddleware) {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		fmt.Println("Creating single instance now.")
		singleInstance = userController
	}
}
