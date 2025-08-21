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
	SendMessage(uid, msg string) error
	SendEvent(uid, eventName, msg string) error
	GetOrCreateChannel(subject string) *NotificationChannel
}

var lock = &sync.Mutex{}
var singleInstance NotificationService

func GetServiceInstance() NotificationService {
	lock.Lock()
	defer lock.Unlock()
	if singleInstance == nil {
		singleInstance = &NotificationDispatcher{
			UserNotifications: make(map[string]*NotificationChannel),
			UserLock:          sync.Mutex{},
			templates:         map[string]string{},
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
