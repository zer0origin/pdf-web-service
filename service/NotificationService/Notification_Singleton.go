package NotificationService

import (
	"fmt"
	"sync"
)

type NotificationService interface {
	CreateNotificationChannel(uid string) *NotificationChannel
	DeleteNotificationChannel(uid string) bool
	GetNotificationChannel(uid string) (*NotificationChannel, error)
	Broadcast(msg string)
}

var lock = &sync.Mutex{}
var singleInstance NotificationService

func GetServiceInstance() NotificationService {
	lock.Lock()
	defer lock.Unlock()
	if singleInstance == nil {
		lock.Lock()
		singleInstance = &NotificationDispatcher{
			UserNotifications: make(map[string]*NotificationChannel),
			UserLock:          sync.Mutex{},
		}
	}

	return singleInstance
}

func SetServiceInstance(NotificationService NotificationService) {
	lock.Lock()
	defer lock.Unlock()
	if singleInstance == nil {
		fmt.Println("Creating single instance now.")
		singleInstance = NotificationService
	}
}
