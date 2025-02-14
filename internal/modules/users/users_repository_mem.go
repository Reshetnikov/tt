package users

import (
	"errors"
	"sync"
)

type UsersRepositoryMem struct {
	mu     sync.Mutex
	users  map[int]*User
	nextID int
}

func NewUsersRepositoryMem() *UsersRepositoryMem {
	return &UsersRepositoryMem{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

func (repo *UsersRepositoryMem) Create(user *User) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	user.ID = repo.nextID
	repo.users[repo.nextID] = user
	repo.nextID++
	return nil
}

func (repo *UsersRepositoryMem) GetByID(id int) *User {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	user, exists := repo.users[id]
	if !exists {
		return nil
	}
	return user
}

func (repo *UsersRepositoryMem) GetByEmail(email string) *User {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, user := range repo.users {
		if user.Email == email {
			return user
		}
	}
	return nil
}

func (repo *UsersRepositoryMem) GetByActivationHash(activationHash string) *User {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, user := range repo.users {
		if user.ActivationHash == activationHash {
			return user
		}
	}
	return nil
}

func (repo *UsersRepositoryMem) Update(user *User) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.users[user.ID]; !exists {
		return errors.New("user not found")
	}
	repo.users[user.ID] = user
	return nil
}

func (repo *UsersRepositoryMem) Delete(id int) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.users[id]; !exists {
		return errors.New("user not found")
	}
	delete(repo.users, id)
	return nil
}
