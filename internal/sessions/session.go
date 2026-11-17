package sessions

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/auth"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{DB: db} }

func (s *Store) Create(userID int64, ttl time.Duration) (token, csrf string, err error) {
	token = randomToken(32)
	csrf = randomToken(16)
	now := time.Now().UTC()
	exp := now.Add(ttl)
	_, err = s.DB.Exec(
		`INSERT INTO sessions (token,user_id,csrf,created_at,expires_at) VALUES (?,?,?,?,?)`,
		token, userID, csrf, now.Format(time.RFC3339), exp.Format(time.RFC3339),
	)
	if err != nil {
		return "", "", err
	}
	return token, csrf, nil
}

func (s *Store) Get(token string) (userID int64, csrf string, err error) {
	var exp string
	err = s.DB.QueryRow(
		`SELECT user_id,csrf,expires_at FROM sessions WHERE token=?`, token,
	).Scan(&userID, &csrf, &exp)
	if err != nil {
		return 0, "", err
	}
	t, err := time.Parse(time.RFC3339, exp)
	if err != nil || time.Now().After(t) {
		s.Delete(token)
		return 0, "", fmt.Errorf("session expired")
	}
	return userID, csrf, nil
}

func (s *Store) Delete(token string) error {
	_, err := s.DB.Exec(`DELETE FROM sessions WHERE token=?`, token)
	return err
}

func (s *Store) DeleteExpired() error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.DB.Exec(`DELETE FROM sessions WHERE expires_at < ?`, now)
	return err
}

func randomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// CookieName for session.
const CookieName = "vertex_session"

// CSRFHeader expected.
const CSRFHeader = "X-CSRF-Token"

// RoleFromUser just re-export.
var _ = auth.RoleAdmin
