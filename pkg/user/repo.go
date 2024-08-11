package user

import (
	"errors"
	"sync"

	"reddit/pkg/idgenerator"

	"reddit/pkg/hasher"
)

var (
	ErrNoUser       = errors.New("no user with such username")
	ErrBadPass      = errors.New("bad password")
	ErrAlreadyExist = errors.New("already exists")
)

type UserDBRepository interface {
	FindUserByUsernameDB(username string) (*User, error)
	AddNewUserDB(newUser *User) error
}

type UserMemoryRepository struct {
	UserDBRepo  UserDBRepository
	mu          *sync.RWMutex
	generatorID idgenerator.IDGenerator
}

func NewUserMemoryRepository(userDBRepo UserDBRepository, idGenerator idgenerator.IDGenerator) *UserMemoryRepository {
	return &UserMemoryRepository{
		mu:          &sync.RWMutex{},
		UserDBRepo:  userDBRepo,
		generatorID: idGenerator,
	}
}

func (u *UserMemoryRepository) Login(username, password string) (*User, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	loginUser, err := u.UserDBRepo.FindUserByUsernameDB(username)
	if err != nil {
		return nil, err
	}
	hashedPassword, err := hasher.GetHashPassword(password)
	if err != nil {
		return nil, err
	}
	if hashedPassword != loginUser.password {
		return nil, ErrBadPass
	}
	return loginUser, nil
}

func (u *UserMemoryRepository) Register(username, password string) (*User, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	hashedPassword, err := hasher.GetHashPassword(password)
	if err != nil {
		return nil, err
	}
	newUser := newUser(u.generatorID.GenerateID(16), username, hashedPassword)
	err = u.UserDBRepo.AddNewUserDB(newUser)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}
