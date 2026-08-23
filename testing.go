package telejoon

import (
	"sync"

	"github.com/aliforever/go-telegram-bot-api/structs"
)

// MockUserRepository is a mock implementation of UserRepository for testing.
// It tracks all method calls for assertions in tests.
type MockUserRepository struct {
	users  sync.Map
	states sync.Map

	// Call tracking for assertions
	UpsertUserCalls []UpsertUserCall
	SetStateCalls   []SetStateCall
	GetStateCalls   []GetStateCall

	mu sync.Mutex
}

// UpsertUserCall records a call to UpsertUser
type UpsertUserCall struct {
	User *structs.User
}

// SetStateCall records a call to SetUserState
type SetStateCall struct {
	UserID int64
	State  string
}

// GetStateCall records a call to GetUserState
type GetStateCall struct {
	UserID int64
}

// NewMockUserRepository creates a new mock user repository for testing
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
}

func (m *MockUserRepository) UpsertUser(user *structs.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UpsertUserCalls = append(m.UpsertUserCalls, UpsertUserCall{User: user})
	m.users.Store(user.Id, user)
	return nil
}

func (m *MockUserRepository) SetUserState(id int64, state string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SetStateCalls = append(m.SetStateCalls, SetStateCall{UserID: id, State: state})
	m.states.Store(id, state)
	return nil
}

func (m *MockUserRepository) GetUserState(id int64) (string, error) {
	m.mu.Lock()
	m.GetStateCalls = append(m.GetStateCalls, GetStateCall{UserID: id})
	m.mu.Unlock()

	if state, ok := m.states.Load(id); ok {
		return state.(string), nil
	}
	return "", UserStateNotFoundErr
}

// SetState pre-sets a state for testing purposes (doesn't track the call)
func (m *MockUserRepository) SetState(userID int64, state string) {
	m.states.Store(userID, state)
}

// SetUser pre-sets a user for testing purposes (doesn't track the call)
func (m *MockUserRepository) SetUser(user *structs.User) {
	m.users.Store(user.Id, user)
}

// GetUser retrieves a user by ID
func (m *MockUserRepository) GetUser(id int64) *structs.User {
	if user, ok := m.users.Load(id); ok {
		return user.(*structs.User)
	}
	return nil
}

// Reset clears all tracked calls and stored data
func (m *MockUserRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UpsertUserCalls = nil
	m.SetStateCalls = nil
	m.GetStateCalls = nil

	// Clear in place instead of replacing the maps: SetState/SetUser/GetUser
	// access them without m.mu, so swapping in a new sync.Map would race.
	m.users.Range(func(key, _ interface{}) bool {
		m.users.Delete(key)
		return true
	})
	m.states.Range(func(key, _ interface{}) bool {
		m.states.Delete(key)
		return true
	})
}

// AssertUpsertUserCalled returns true if UpsertUser was called for the given user ID
func (m *MockUserRepository) AssertUpsertUserCalled(userID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, call := range m.UpsertUserCalls {
		if call.User.Id == userID {
			return true
		}
	}
	return false
}

// AssertStateSet returns true if SetUserState was called with the given user ID and state
func (m *MockUserRepository) AssertStateSet(userID int64, state string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, call := range m.SetStateCalls {
		if call.UserID == userID && call.State == state {
			return true
		}
	}
	return false
}

// MockUserLanguageRepository is a mock implementation of UserLanguageRepository for testing
type MockUserLanguageRepository struct {
	languages sync.Map
	mu        sync.Mutex

	SetLanguageCalls []SetLanguageCall
	GetLanguageCalls []GetLanguageCall
}

// SetLanguageCall records a call to SetUserLanguage
type SetLanguageCall struct {
	UserID   int64
	Language string
}

// GetLanguageCall records a call to GetUserLanguage
type GetLanguageCall struct {
	UserID int64
}

// NewMockUserLanguageRepository creates a new mock language repository
func NewMockUserLanguageRepository() *MockUserLanguageRepository {
	return &MockUserLanguageRepository{}
}

func (m *MockUserLanguageRepository) SetUserLanguage(id int64, language string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SetLanguageCalls = append(m.SetLanguageCalls, SetLanguageCall{UserID: id, Language: language})
	m.languages.Store(id, language)
	return nil
}

func (m *MockUserLanguageRepository) GetUserLanguage(id int64) (string, error) {
	m.mu.Lock()
	m.GetLanguageCalls = append(m.GetLanguageCalls, GetLanguageCall{UserID: id})
	m.mu.Unlock()

	if lang, ok := m.languages.Load(id); ok {
		return lang.(string), nil
	}
	return "", UserLanguageNotFoundErr
}

// SetLanguage pre-sets a language for testing
func (m *MockUserLanguageRepository) SetLanguage(userID int64, language string) {
	m.languages.Store(userID, language)
}

// Reset clears all tracked calls and stored data
func (m *MockUserLanguageRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SetLanguageCalls = nil
	m.GetLanguageCalls = nil

	// Clear in place instead of replacing the map (see MockUserRepository.Reset).
	m.languages.Range(func(key, _ interface{}) bool {
		m.languages.Delete(key)
		return true
	})
}
