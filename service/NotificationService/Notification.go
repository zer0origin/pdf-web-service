package NotificationService

import (
	"errors"
	"fmt"
	"sync"
)

type NotificationService interface {
	CreateNotificationChannel(uid string) *NotificationChannel
	DeleteNotificationChannel(uid string) bool
	GetNotificationChannel(uid string) (*NotificationChannel, error)
	Broadcast(msg string)
}

type NotificationDispatcher struct {
	UserNotifications map[string]*NotificationChannel
	UserLock          sync.Mutex
}

type NotificationChannel struct {
	Channel          chan string
	ConnectedClients int
}

var lock = &sync.Mutex{}
var singleInstance NotificationService

func GetInstance() NotificationService {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		singleInstance = &NotificationDispatcher{
			UserNotifications: make(map[string]*NotificationChannel),
			UserLock:          sync.Mutex{},
		}
	}

	return singleInstance
}

func SetInstance(NotificationService NotificationService) {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		fmt.Println("Creating single instance now.")
		singleInstance = NotificationService
	}
}

func (t *NotificationDispatcher) CreateNotificationChannel(uid string) *NotificationChannel {
	t.UserLock.Lock()
	defer t.UserLock.Unlock()

	notifChannel := &NotificationChannel{Channel: make(chan string), ConnectedClients: 1}
	t.UserNotifications[uid] = notifChannel
	fmt.Println("Created new channel for user " + uid)
	return notifChannel
}

func (t *NotificationDispatcher) GetNotificationChannel(uid string) (*NotificationChannel, error) {
	t.UserLock.Lock()
	defer t.UserLock.Unlock()

	ch, ok := t.UserNotifications[uid]
	if !ok {
		return &NotificationChannel{}, errors.New("user does not have an open notification channel")
	}

	fmt.Println("Reused channel for user " + uid)
	return ch, nil
}

func (t *NotificationDispatcher) DeleteNotificationChannel(uid string) bool {
	t.UserLock.Lock()
	defer t.UserLock.Unlock()

	notificationChannel, ok := t.UserNotifications[uid]
	if ok {
		if notificationChannel.ConnectedClients == 1 {
			delete(t.UserNotifications, uid)
			return true
		}

		notificationChannel.ConnectedClients = notificationChannel.ConnectedClients - 1
	}

	return false
}

func (t *NotificationDispatcher) Broadcast(msg string) {
	for uid, ch := range t.UserNotifications {
		fmt.Println("Sent message to " + uid)
		for range ch.ConnectedClients {
			ch.Channel <- msg
		}
	}
}
