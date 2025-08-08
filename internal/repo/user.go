package repo

import (
	"strings"
	"time"

	"github.com/NOTMKW/JWT/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(user *model.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	return r.db.Create(user).Error
}

func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByGoogleID(googleID string) (*model.User, error) {
	var user model.User
	err := r.db.Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) StoreMFACode(mfaCode *model.MFACode) error {
	r.db.Where("email = ?", mfaCode.Email).Delete(&model.MFACode{})

	return r.db.Create(mfaCode).Error
}

func (r *UserRepository) GetMFACode(email string) (*model.MFACode, error) {
	var mfaCode model.MFACode
	err := r.db.Where("email = ? AND expires_at > ?", email, time.Now()).First(&mfaCode).Error
	if err != nil {
		return nil, err
	}
	return &mfaCode, nil
}

func (r *UserRepository) DeleteMFACode(email string) error {
	return r.db.Where("email = ?", email).Delete(&model.MFACode{}).Error
}

func (r *UserRepository) CreateOrUpdateGoogleUser(googleID, email, name string) (*model.User, error) {
	user, err := r.GetUserByGoogleID(googleID)
	if err != nil {
		return user, nil
	}

	user, err = r.GetUserByEmail(email)
	if err == nil {
		user.GoogleID = googleID
		r.db.Save(user)
		return user, nil
	}

	user = &model.User{
		ID:       uuid.New().String(),
		Username: r.generateUsernameFromEmail(email),
		Email:    email,
		Role:     model.RoleUser,
		GoogleID: googleID,
	}

	err = r.db.Create(user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) generateUsernameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return "user"
}
