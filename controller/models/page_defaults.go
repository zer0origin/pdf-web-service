package models

type PageDefaults struct {
	NavDetails           NavDetails
	NotificationSettings *NotificationSettings
	ContentDetails       any
}

type NavDetails struct {
	IsAuthenticated bool
}

type NotificationSettings struct {
	Uid string
}
