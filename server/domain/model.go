package domain

import (
	"time"
)

type ValueType int

const (
	TypeString ValueType = iota
	TypeList
)

// Entry holds data and an expiry timestamp.
type Entry struct {
	Type   ValueType
	Str    string
	Items  []string
	Expiry time.Time
}

// NewStringEntry creates a string entry that expires after ttl.
func NewStringEntry(str string, expiry time.Duration) *Entry {

	return &Entry{
		Type:   TypeString,
		Str:    str,
		Expiry: time.Now().Add(expiry),
	}

}

// NewListEntry creates a list entry initialized with items and a TTL.
func NewListEntry(items []string, expiry time.Duration) *Entry {

	return &Entry{
		Type:   TypeList,
		Items:  append([]string(nil), items...), // clone slice
		Expiry: time.Now().Add(expiry),
	}
}
