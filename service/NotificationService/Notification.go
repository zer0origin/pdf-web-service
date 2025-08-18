package NotificationService

import (
	"errors"
	"fmt"
	"sync"
)

type NotificationService interface {
	CreateNotificationChannel(uid string) chan string
	DeleteNotificationChannel(uid string) bool
	GetNotificationChannel(uid string) (chan string, error)
	Broadcast(msg string)
}

type NotificationDispatcher struct {
	UserNotifications map[string]chan string
	UserLock          sync.Mutex
}

var singleInstance NotificationService

var lock = &sync.Mutex{}

func GetInstance() NotificationService {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		fmt.Println("Creating single instance now.")
		singleInstance = &NotificationDispatcher{
			UserNotifications: map[string]chan string{},
			UserLock:          sync.Mutex{},
		}
	}

	return singleInstance
}

func (t *NotificationDispatcher) CreateNotificationChannel(uid string) chan string {
	t.UserLock.Lock()
	defer t.UserLock.Unlock()

	userChannel := make(chan string)
	t.UserNotifications[uid] = userChannel
	return userChannel
}

func (t *NotificationDispatcher) GetNotificationChannel(uid string) (chan string, error) {
	ch, ok := t.UserNotifications[uid]

	if ok {
		return ch, nil
	}

	return nil, errors.New("user does not have an open notification channel")
}

func (t *NotificationDispatcher) DeleteNotificationChannel(uid string) bool {
	t.UserLock.Lock()
	defer t.UserLock.Lock()

	_, ok := t.UserNotifications[uid]

	if ok {
		delete(t.UserNotifications, uid)
		return true
	}

	return false
}

func (t *NotificationDispatcher) Broadcast(msg string) {
	for _, ch := range t.UserNotifications {
		ch <- msg
	}
}
