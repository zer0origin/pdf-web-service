package models

type PageDefaults struct {
	NavDetails           *NavDetails
	NotificationSettings *NotificationSettings
	ContentDetails       any
}

type NavDetails struct {
	IsAuthenticated bool
}

type NotificationSettings struct {
	Uid string
}

type PageInfo struct {
	Offset   uint32
	NextPage *uint32
	LastPage *uint32
	Limit    uint32
}
