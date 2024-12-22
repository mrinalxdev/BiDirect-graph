package models

type MemberID int64

type Connection struct {
	SourceID MemberID
	DestID MemberID
	Metadata map[string]interface{}
}

type GraphDistance struct {
	SourceID MemberID
	DestID MemberID
	Distance int
}
