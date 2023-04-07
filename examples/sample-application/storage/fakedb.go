// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build go1.16
// +build go1.16

package storage

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"

	"golang.org/x/crypto/scrypt"
)

// Note: a real program would connect to a real DB using
// github.com/google/go-safeweb/safesql. This is just a simple storage
// implementation to demonstrate the web framework.

// Please ignore the contents of this file.

type Note struct {
	Title, Text string
}

type DB struct {
	mu sync.Mutex
	// user -> note title -> notes
	notes map[string]map[string]Note

	// user -> token
	sessionTokens map[string]string
	// token -> user
	userSessions map[string]string

	// user -> pw hash
	credentials map[string]string
}

func NewDB() *DB {
	return &DB{
		notes:         map[string]map[string]Note{},
		sessionTokens: map[string]string{},
		userSessions:  map[string]string{},
		credentials:   map[string]string{},
	}
}

// Notes

func (s *DB) AddOrEditNote(user string, n Note) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.notes[user] == nil {
		s.notes[user] = map[string]Note{}
	}
	s.notes[user][n.Title] = n
}

func (s *DB) GetNotes(user string) []Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	var ns []Note
	for _, n := range s.notes[user] {
		ns = append(ns, n)
	}
	return ns
}

// Sessions

func (s *DB) GetUser(token string) (user string, valid bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, valid = s.sessionTokens[token]
	return user, valid
}

func (s *DB) GetToken(user string) (token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, has := s.userSessions[user]
	if has {
		return token
	}
	token = genToken()
	s.userSessions[user] = token
	s.sessionTokens[token] = user
	return token
}

func (s *DB) DelSession(user string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, has := s.userSessions[user]
	if !has {
		return
	}
	delete(s.userSessions, user)
	delete(s.sessionTokens, token)
}

func genToken() string {
	b := make([]byte, 20)
	rand.Read(b)
	tok := base64.RawStdEncoding.EncodeToString(b)
	return tok
}

// Credentials

// HasUser checks if the user exists.
func (s *DB) HasUser(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, has := s.credentials[name]
	return has
}

// AddUser adds a user to the storage if it is not already there.
func (s *DB) AddOrAuthUser(name, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if password == "" {
		return errors.New("password cannot be empty")
	}
	if storedHash, has := s.credentials[name]; has {
		if storedHash != hash(password) {
			return errors.New("wrong password")
		}
		return nil
	}
	s.credentials[name] = hash(password)
	return nil
}

func hash(pw string) string {
	salt := []byte("please use a proper salt in production")
	hash, err := scrypt.Key([]byte(pw), salt, 32768, 8, 1, 32)
	if err != nil {
		panic("this should not happen")
	}
	return string(hash)
}
