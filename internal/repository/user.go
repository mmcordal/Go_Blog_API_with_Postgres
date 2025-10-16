package repository

import (
	"cleanArch_with_postgres/internal/entity"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, oldUsername string, user *entity.User) error
	Delete(ctx context.Context, username string) error
	ExistUser(ctx context.Context, email, username string) (bool, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByIdentifier(ctx context.Context, identifier string) (*entity.User, error)
	SearchByUsernamePrefix(ctx context.Context, prefix string, limit int) ([]entity.User, error)
	SearchByUsernamePrefixWithOptions(ctx context.Context, prefix string, limit int, includeDeleted bool) ([]entity.User, error)
	Restore(ctx context.Context, username string) error
	SetRole(ctx context.Context, username string, role entity.UserRole) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		fmt.Println("user create error:", err)
		return err
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, oldUsername string, user *entity.User) error {
	err := r.db.WithContext(ctx).Model(user).
		Where("username = ?", oldUsername).
		Updates(map[string]interface{}{
			"username":   user.Username,
			"email":      user.Email,
			"password":   user.Password,
			"updated_at": time.Now(),
		}).Error

	if err != nil {
		fmt.Println("user update error:", err)
		return err
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, username string) error {
	user, err := r.GetByUsername(ctx, username)
	if err != nil {
		fmt.Println("user delete error:", err)
		return err
	}
	err = r.db.WithContext(ctx).Model(user).
		Where("username = ?", username).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
		}).Error

	if err != nil {
		fmt.Println("user delete error2:", err)
		return err
	}
	return nil
}

func (r *userRepository) ExistUser(ctx context.Context, email, username string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("email = ?", email).
		Count(&count).Error; err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	if err := r.db.
		Model(&entity.User{}).WithContext(ctx).
		Where("username = ?", username).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User

	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByIdentifier(ctx context.Context, identifier string) (*entity.User, error) {
	var user entity.User

	err := r.db.WithContext(ctx).
		Where("email = ? OR username = ?", identifier, identifier).
		First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) SearchByUsernamePrefix(ctx context.Context, prefix string, limit int) ([]entity.User, error) {
	var users []entity.User
	if limit <= 0 {
		limit = 10
	}
	err := r.db.WithContext(ctx).
		Where("username ILIKE ?", prefix+"%").
		Limit(limit).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) SearchByUsernamePrefixWithOptions(ctx context.Context, prefix string, limit int, includeDeleted bool) ([]entity.User, error) {
	var users []entity.User
	if limit <= 0 {
		limit = 10
	}

	q := r.db.WithContext(ctx)
	if includeDeleted {
		q = q.Unscoped() // soft-deleted dahil
	}

	if err := q.
		Where("username ILIKE ?", prefix+"%").
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Restore(ctx context.Context, username string) error {
	tx := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Unscoped(). // deleted_at NULL yapabilmek için
		Where("username = ?", username).
		Updates(map[string]interface{}{
			"deleted_at": nil,
			"updated_at": time.Now(),
		})

	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *userRepository) SetRole(ctx context.Context, username string, role entity.UserRole) error {
	return r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("username = ?", username).
		Updates(map[string]interface{}{
			"role":       role,
			"updated_at": time.Now(),
		}).Error
}
