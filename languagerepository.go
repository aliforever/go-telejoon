package telejoon

import (
	"errors"
	"sync"
)

var UserLanguageNotFoundErr = errors.New("user_language_not_found")

type UserLanguageRepository interface {
	SetUserLanguage(userID int64, languageTag string) error
	GetUserLanguage(userID int64) (string, error)
}

// DefaultUserLanguageRepository is a default implementation of UserLanguageRepository.
type DefaultUserLanguageRepository struct {
	languages sync.Map
}

// NewDefaultUserLanguageRepository creates a new DefaultUserLanguageRepository.
func NewDefaultUserLanguageRepository() UserLanguageRepository {
	return &DefaultUserLanguageRepository{
		languages: sync.Map{},
	}
}

// SetUserLanguage sets a language for a user.
func (u *DefaultUserLanguageRepository) SetUserLanguage(userID int64, languageTag string) error {
	u.languages.Store(userID, languageTag)
	return nil
}

// GetUserLanguage gets a language for a user.
func (u *DefaultUserLanguageRepository) GetUserLanguage(userID int64) (string, error) {
	if language, ok := u.languages.Load(userID); ok {
		return language.(string), nil
	}

	return "", nil
}
