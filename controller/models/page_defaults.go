package models

type PageDefaults struct {
	NavDetails     NavDetails
	ContentDetails any
}

type NavDetails struct {
	IsAuthenticated bool
}
