package users

import (
	"database/sql"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/auth"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{DB: db} }

func (s *Store) Create(username, password string, role auth.Role) (auth.User, error) {
	hash, err := auth.HashPassword(password, auth.DefaultParams)
	if err != nil {
		return auth.User{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.DB.Exec(
		`INSERT INTO users (username,password_hash,role,created_at) VALUES (?,?,?,?)`,
		username, hash, string(role), now,
	)
	if err != nil {
		return auth.User{}, err
	}
	id, _ := res.LastInsertId()
	return auth.User{ID: id, Username: username, PasswordHash: hash, Role: role, CreatedAt: parseTime(now)}, nil
}

func (s *Store) GetByUsername(username string) (auth.User, error) {
	row := s.DB.QueryRow(
		`SELECT id,username,password_hash,role,created_at,failed_attempts,locked_until FROM users WHERE username=?`,
		username,
	)
	return scanUser(row)
}

func (s *Store) GetByID(id int64) (auth.User, error) {
	row := s.DB.QueryRow(
		`SELECT id,username,password_hash,role,created_at,failed_attempts,locked_until FROM users WHERE id=?`,
		id,
	)
	return scanUser(row)
}

func (s *Store) List() ([]auth.User, error) {
	rows, err := s.DB.Query(`SELECT id,username,password_hash,role,created_at,failed_attempts,locked_until FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []auth.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Store) UpdatePassword(id int64, password string) error {
	hash, err := auth.HashPassword(password, auth.DefaultParams)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`UPDATE users SET password_hash=? WHERE id=?`, hash, id)
	return err
}

func (s *Store) UpdateRole(id int64, role auth.Role) error {
	_, err := s.DB.Exec(`UPDATE users SET role=? WHERE id=?`, string(role), id)
	return err
}

func (s *Store) Delete(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM users WHERE id=?`, id)
	return err
}

func (s *Store) RecordFailedAttempt(u auth.User) error {
	attempts := u.FailedAttempts + 1
	var lockedUntil *string
	if attempts >= 5 {
		t := time.Now().Add(15 * time.Minute).UTC().Format(time.RFC3339)
		lockedUntil = &t
		attempts = 0
	}
	if lockedUntil != nil {
		_, err := s.DB.Exec(`UPDATE users SET failed_attempts=?, locked_until=? WHERE id=?`, attempts, *lockedUntil, u.ID)
		return err
	}
	_, err := s.DB.Exec(`UPDATE users SET failed_attempts=? WHERE id=?`, attempts, u.ID)
	return err
}

func (s *Store) ResetFailedAttempts(id int64) error {
	_, err := s.DB.Exec(`UPDATE users SET failed_attempts=0, locked_until=NULL WHERE id=?`, id)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(s scanner) (auth.User, error) {
	var u auth.User
	var role string
	var created string
	var locked sql.NullString
	var failed int
	if err := s.Scan(&u.ID, &u.Username, &u.PasswordHash, &role, &created, &failed, &locked); err != nil {
		return auth.User{}, err
	}
	u.Role = auth.Role(role)
	u.CreatedAt = parseTime(created)
	u.FailedAttempts = failed
	if locked.Valid {
		t := parseTime(locked.String)
		u.LockedUntil = &t
	}
	return u, nil
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
