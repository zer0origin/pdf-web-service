package NotificationService

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type NotificationDispatcher struct {
	UserNotifications map[string]*NotificationChannel
	UserLock          sync.Mutex
	templates         map[string]string
} //TODO: We should keep track of issues uids, or maybe a JWT token.

type NotificationChannel struct {
	Channel          chan string
	ConnectedClients int
}

func (t *NotificationDispatcher) CreateNotificationChannel(uid string) *NotificationChannel {
	t.UserLock.Lock()
	defer t.UserLock.Unlock()

	notifChannel := &NotificationChannel{Channel: make(chan string, 1), ConnectedClients: 1}
	t.UserNotifications[uid] = notifChannel
	fmt.Println("[NotificationService] Created new channel for user " + uid)
	return notifChannel
}

func (t *NotificationDispatcher) GetOrCreateNotificationChannel(uid string) (*NotificationChannel, error) {
	t.UserLock.Lock()
	defer t.UserLock.Unlock()

	ch, ok := t.UserNotifications[uid]
	if !ok {
		notifChannel := &NotificationChannel{Channel: make(chan string, 1), ConnectedClients: 1}
		t.UserNotifications[uid] = notifChannel
		fmt.Println("[NotificationService] Created new channel for uid " + uid)
		return notifChannel, nil
	}

	ch.ConnectedClients += 1
	fmt.Println(fmt.Sprintf("[NotificationService] Reused channel for user %s | Connected %d", uid, ch.ConnectedClients))
	return ch, nil
}

func (t *NotificationDispatcher) GetNotificationChannel(uid string) (*NotificationChannel, error) {
	t.UserLock.Lock()
	defer t.UserLock.Unlock()

	ch, ok := t.UserNotifications[uid]
	if !ok {
		return &NotificationChannel{}, errors.New("user does not have an open notification channel")
	}

	fmt.Println("[NotificationService] Reused channel for user " + uid)
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
	for uid := range t.UserNotifications {
		_ = t.SendMessage(uid, msg)
	}
}

func (t *NotificationDispatcher) SendMessage(uid, msg string) error {
	notificationChannel, ok := t.UserNotifications[uid]
	if !ok {
		return NoNotificationChannel
	}

	fmt.Printf("[NotificationService] Sent message to %s | Connected %d\n", uid, notificationChannel.ConnectedClients)
	for range notificationChannel.ConnectedClients {
		notificationChannel.Channel <- fmt.Sprintf("data: %s\n\n", msg)
	}

	return nil
}

var NoNotificationChannel = errors.New("cannot find users notification channel")

func (t *NotificationDispatcher) SendEvent(uid, eventName, msg string) error {
	notificationChannel, ok := t.UserNotifications[uid]
	if !ok {
		return NoNotificationChannel
	}

	fmt.Printf("[NotificationService] Sent event %s to %s | Connected %d\n", eventName, uid, notificationChannel.ConnectedClients)
	for range notificationChannel.ConnectedClients {
		notificationChannel.Channel <- fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, msg)
	}

	return nil
}

// SendEventToAllInstancesOfUser send an event to all connected instances of a user. Only possible if grouped uid is used, (i.e. "Subject.InstanceIdentifier")
func (t *NotificationDispatcher) SendEventToAllInstancesOfUser(uid, eventName, msg string) {
	if strings.Contains(uid, ".") {
		subject := strings.Split(uid, ".")[0]

		for uidStored, _ := range t.UserNotifications {
			if strings.Contains(uidStored, ".") {
				uidStoredSubject := strings.Split(uidStored, ".")[0]
				if subject == uidStoredSubject {
					_ = t.SendEvent(uidStored, eventName, msg)
				}
			}
		}

		return
	}

	_ = t.SendEvent(uid, eventName, msg)
}
