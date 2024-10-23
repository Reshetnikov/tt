package users

import (
	"errors"
	"sync"
)

type UserRepositoryMem struct {
	mu     sync.Mutex
	users  map[int]*User
	nextID int
}

func NewUserRepositoryMem() *UserRepositoryMem {
	return &UserRepositoryMem{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

func (repo *UserRepositoryMem) Create(user *User) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	user.ID = repo.nextID
	repo.users[repo.nextID] = user
	repo.nextID++
	return nil
}

func (repo *UserRepositoryMem) GetByID(id int) (*User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	user, exists := repo.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (repo *UserRepositoryMem) GetByEmail(email string) (*User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, user := range repo.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (repo *UserRepositoryMem) Update(user *User) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.users[user.ID]; !exists {
		return errors.New("user not found")
	}
	repo.users[user.ID] = user
	return nil
}

func (repo *UserRepositoryMem) Delete(id int) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.users[id]; !exists {
		return errors.New("user not found")
	}
	delete(repo.users, id)
	return nil
}
