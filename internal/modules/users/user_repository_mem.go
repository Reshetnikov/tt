package users

import (
	"errors"
	"sync"
)

// UserRepositoryMem is an in-memory implementation of UserRepository
type UserRepositoryMem struct {
	mu     sync.Mutex
	users  map[int]*User
	nextID int
}

// UserRepositoryMem creates a new instance of UserRepositoryMem
func NewUserRepositoryMem() *UserRepositoryMem {
	return &UserRepositoryMem{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

// Create adds a new user to the in-memory store
func (repo *UserRepositoryMem) Create(user *User) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	user.ID = repo.nextID
	repo.users[repo.nextID] = user
	repo.nextID++
	return nil
}

// GetByID retrieves a user by their ID
func (repo *UserRepositoryMem) GetByID(id int) (*User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	user, exists := repo.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetByUsername retrieves a user by their username
func (repo *UserRepositoryMem) GetByUsername(username string) (*User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, user := range repo.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

// Update updates an existing user in the in-memory store
func (repo *UserRepositoryMem) Update(user *User) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.users[user.ID]; !exists {
		return errors.New("user not found")
	}
	repo.users[user.ID] = user
	return nil
}

// Delete removes a user from the in-memory store
func (repo *UserRepositoryMem) Delete(id int) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.users[id]; !exists {
		return errors.New("user not found")
	}
	delete(repo.users, id)
	return nil
}
