package paste

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Store is a thread-safe in-memory paste store.
type Store struct {
	mu     sync.RWMutex
	pastes map[string]Paste
}

// NewStore returns an empty Store.
func NewStore() *Store {
	return &Store{
		pastes: make(map[string]Paste),
	}
}

// Create stores a new paste and returns it. The id is 16 random bytes hex
// encoded (32 characters). If expiresIn is greater than zero, ExpiresAt is set
// to CreatedAt + expiresIn; otherwise the paste never expires.
func (s *Store) Create(content, language string, expiresIn time.Duration) (Paste, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return Paste{}, err
	}
	id := hex.EncodeToString(b)

	now := time.Now()
	p := Paste{
		ID:        id,
		Content:   content,
		Language:  language,
		CreatedAt: now,
	}
	if expiresIn > 0 {
		e := now.Add(expiresIn)
		p.ExpiresAt = &e
	}

	s.mu.Lock()
	s.pastes[id] = p
	s.mu.Unlock()

	return p, nil
}

// Get returns the paste with the given id. The second return value is false
// when the id is unknown or the paste has expired.
func (s *Store) Get(id string) (Paste, bool) {
	s.mu.RLock()
	p, ok := s.pastes[id]
	s.mu.RUnlock()
	if !ok {
		return Paste{}, false
	}
	if s.IsExpired(p) {
		return Paste{}, false
	}
	return p, true
}

// List returns all non-expired pastes.
func (s *Store) List() []Paste {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Paste, 0, len(s.pastes))
	for _, p := range s.pastes {
		if !s.IsExpired(p) {
			out = append(out, p)
		}
	}
	return out
}

// Delete removes the paste with the given id and reports whether it existed.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.pastes[id]; !ok {
		return false
	}
	delete(s.pastes, id)
	return true
}

// IsExpired reports whether p has an expiry time in the past.
func (s *Store) IsExpired(p Paste) bool {
	if p.ExpiresAt == nil {
		return false
	}
	return !p.ExpiresAt.After(time.Now())
}
