package store_service

import (
	"context"
	domain2 "data_storage/server/domain"
	"errors"
	"fmt"
	"time"
)

// StoreServiceRepo abstracts adapter operations.
type StoreServiceRepo interface {
	SetString(ctx context.Context, key string, data string, ttl time.Duration) error
	GetString(ctx context.Context, key string) (string, error)
	DeleteString(ctx context.Context, key string) error

	LPush(ctx context.Context, key string, items ...string) error
	RPop(ctx context.Context, key string) (string, error)
}

// StoreService implements business logic.
type StoreService struct {
	domainRepo domain2.EntryRepository
	defaultTTL time.Duration
}

// NewStoreService wires repo + default TTL.
func NewStoreService(d domain2.EntryRepository, defaultTTL time.Duration) StoreServiceRepo {
	return &StoreService{
		domainRepo: d,
		defaultTTL: defaultTTL,
	}
}

// SetString validates inputs and stores a string.
func (s *StoreService) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {

	if key == "" {
		err := domain2.ErrEmptyKey
		return fmt.Errorf("SetString: %q: %w", key, err)
	}

	if value == "" {
		return fmt.Errorf("SetString %q: %w", key, domain2.ErrEmptyValue)
	}

	expiry := ttl

	if ttl == 0 {
		expiry = s.defaultTTL
	}

	entry := domain2.NewStringEntry(value, expiry)

	if err := s.domainRepo.Set(ctx, key, entry); err != nil {
		return fmt.Errorf("SetString %q: %w", key, err)
	}

	return nil
}

// GetString retrieves a string, erroring on missing/expired/wrong type.
func (s *StoreService) GetString(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("GetString: %q: %w", key, domain2.ErrEmptyKey)
	}
	entry, err := s.domainRepo.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("GetString %q: %w", key, err)
	}
	if entry.Type != domain2.TypeString {
		return "", fmt.Errorf("GetString: %q: %w", key, domain2.ErrWrongType)
	}

	return entry.Str, nil
}

// DeleteString removes a key.
func (s *StoreService) DeleteString(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("DeleteString: %q: %w", key, domain2.ErrEmptyKey)
	}

	err := s.domainRepo.Remove(ctx, key)
	if err != nil {
		return fmt.Errorf("DeleteString: %q: %w", key, err)
	}

	return nil
}

// LPush pushes items onto list head.
func (s *StoreService) LPush(ctx context.Context, key string, items ...string) error {
	if key == "" {
		return fmt.Errorf("LPush: %q: %w", key, domain2.ErrEmptyKey)
	}

	existingList, err := s.domainRepo.Get(ctx, key)
	if errors.Is(err, domain2.ErrNotFound) {
		existingList = domain2.NewListEntry(items, s.defaultTTL)
	} else if err != nil {
		return fmt.Errorf("LPush: %q: %w", key, err)
	} else {
		if existingList.Type != domain2.TypeList {
			return fmt.Errorf("LPush: %q: %w", key, domain2.ErrWrongType)
		}
		existingList.Items = append(items, existingList.Items...)
	}

	if err = s.domainRepo.Set(ctx, key, existingList); err != nil {
		return fmt.Errorf("LPush %q: %w", key, err)
	}
	return nil
}

// RPop pops an item from list tail.
func (s *StoreService) RPop(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("RPop: %q: %w", key, domain2.ErrEmptyKey)
	}

	existingList, err := s.domainRepo.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("RPop %q: %w", key, err)
	}

	if existingList.Type != domain2.TypeList {
		return "", fmt.Errorf("RPop: %q: %w", key, domain2.ErrWrongType)
	}

	n := len(existingList.Items)
	if n == 0 {
		return "", domain2.ErrEmptyEntry
	}

	value := existingList.Items[n-1]
	existingList.Items = existingList.Items[:n-1]

	if err = s.domainRepo.Set(ctx, key, existingList); err != nil {
		return "", fmt.Errorf("RPop: %q: %w", key, err)
	}

	return value, nil

}
