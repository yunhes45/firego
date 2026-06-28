package transfer

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

type FileInfo struct {
	FileID   string
	Filename string
	Status   string
}

type Session struct {
	GroupID   string
	Files     []*FileInfo
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

func (sm *SessionManager) CreateSession(filenames []string, ttl, password string) (*Session, error) {
	duration, err := time.ParseDuration(ttl)

	if err != nil {
		return nil, fmt.Errorf("Invalid TTL Type: %v", err)
	}

	groupID, err := generateCode()

	if err != nil {
		return nil, err
	}

	files := make([]*FileInfo, len(filenames))

	for i, filename := range filenames {
		fileID, err := generateCode()
		if err != nil {
			return nil, err
		}

		files[i] = &FileInfo{
			FileID:   fileID,
			Filename: filename,
			Status:   "pending",
		}
	}

	session := &Session{
		GroupID:   groupID,
		Files:     files,
		TTL:       duration,
		ExpiresAt: time.Now().Add(duration),
		Password:  password,
		CreatedAt: time.Now(),
	}

	sm.mu.Lock()
	sm.sessions[groupID] = session
	sm.mu.Unlock()

	return session, nil
}

func (sm *SessionManager) GetSession(groupID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[groupID]

	if !exists {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, false
	}

	return session, true
}

func generateCode() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%X", b), nil
}
