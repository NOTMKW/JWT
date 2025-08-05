package repo

import (
	"errors"
	"sync"
	"time"

	"github.com/NOTMKW/JWT/internal/model"
	"github.com/google/uuid"
)

type UserRepository struct {
	users map[string]*model.User
	mutex sync.RWMutex
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*model.User),
	}
}

func (r *UserRepository) CreateUser(user *model.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, existingUser := range r.users {
		if existingUser.Email == user.Email {
			return errors.New("email already exists")
		}
		if existingUser.Username == user.Username {
			return errors.New("username already exists")
		}
	}

	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *UserRepository) GetUserByID (id string) (*model.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}
