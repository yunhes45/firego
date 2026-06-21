package transfer

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

type Session struct {
	ID        string
	Filename  string
	TTL       time.Duration
	ExpiresAt time.Time
	Password  string
	CreatedAt time.Time
}

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) CreateSession(filename, ttl, password string) (*Session, error) {
	duration, err := time.ParseDuration(ttl)

	if err != nil {
		return nil, fmt.Errorf("Invalid TTL Type: %v", err)
	}

	id, err := generateCode()

	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:        id,
		Filename:  filename,
		TTL:       duration,
		ExpiresAt: time.Now().Add(duration),
		Password:  password,
		CreatedAt: time.Now(),
	}

	sm.mu.Lock()
	sm.sessions[id] = session
	sm.mu.Unlock()

	return session, nil
}

func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[id]

	if !exists {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, false
	}

	return session, true
}

func generateCode() (string, error) {
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%X", b), nil
}
