package adapters

import (
	"data_storage/server/domain"
	"errors"
)

func isClientError(err error) bool {
	switch {
	case errors.Is(err, domain.ErrEmptyKey),
		errors.Is(err, domain.ErrEmptyValue),
		errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrWrongType),
		errors.Is(err, domain.ErrEmptyEntry),
		errors.Is(err, domain.ErrExpiredEntry):
		return true
	}
	return false
}
