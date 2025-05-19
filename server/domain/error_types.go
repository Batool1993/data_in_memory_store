package domain

import "errors"

var (
	ErrNotFound     = errors.New("entry not found")
	ErrWrongType    = errors.New("wrong entry type")
	ErrEmptyEntry   = errors.New("entry is empty")
	ErrEmptyKey     = errors.New("key or id is empty")
	ErrEmptyValue   = errors.New("value is empty")
	ErrExpiredEntry = errors.New("entry has expired")
)
