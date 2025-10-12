package NotificationService

import (
	"fmt"
	"sync"
)

// TODO: Change notification system to work per session (per token), not per user. client_id for notification param.
type NotificationService interface {
	CreateNotificationChannel(uid string) *NotificationChannel
	DeleteNotificationChannel(uid string) bool
	GetNotificationChannel(uid string) (*NotificationChannel, error)
	Broadcast(msg string)
	SendMessage(uid, msg string) error
	SendEvent(uid, eventName, msg string) error
	GetOrCreateNotificationChannel(uid string) (*NotificationChannel, error)
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
