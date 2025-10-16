package repository

import (
	"cleanArch_with_postgres/internal/entity"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type RoleRequestRepository interface {
	Create(ctx context.Context, r *entity.RoleRequest) error
	LatestByUser(ctx context.Context, username string) (*entity.RoleRequest, error)
	List(ctx context.Context, status entity.RoleRequestStatus, limit int) ([]entity.RoleRequest, error)
	Approve(ctx context.Context, id uint, adminUsername string) error
	Reject(ctx context.Context, id uint, adminUsername string) error
	GetByID(ctx context.Context, id uint) (*entity.RoleRequest, error)
}

type roleRequestRepository struct{ db *gorm.DB }

func NewRoleRequestRepository(db *gorm.DB) RoleRequestRepository {
	return &roleRequestRepository{db: db}
}

func (r *roleRequestRepository) Create(ctx context.Context, rr *entity.RoleRequest) error {
	return r.db.WithContext(ctx).Create(rr).Error
}
func (r *roleRequestRepository) LatestByUser(ctx context.Context, username string) (*entity.RoleRequest, error) {
	var rr entity.RoleRequest
	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		Order("created_at DESC").
		First(&rr).Error
	if err != nil {
		return nil, err
	}
	return &rr, nil
}
func (r *roleRequestRepository) List(ctx context.Context, status entity.RoleRequestStatus, limit int) ([]entity.RoleRequest, error) {
	var rows []entity.RoleRequest
	q := r.db.WithContext(ctx)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if limit <= 0 {
		limit = 100
	}
	if err := q.Order("created_at DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
func (r *roleRequestRepository) Approve(ctx context.Context, id uint, admin string) error {
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&entity.RoleRequest{}).
		Where("id = ? AND status = ?", id, entity.RoleReqPending).
		Updates(map[string]interface{}{"status": entity.RoleReqApproved, "decided_by": admin, "decided_at": &now}).Error
	if err != nil {
		fmt.Println("Error approving role request:", err)
		return err
	} else {
		err = r.db.WithContext(ctx).Model(&entity.User{}).
			Where("username = (SELECT username FROM role_requests WHERE id = ?)", id).
			Update("role", "admin").Error
		if err != nil {
			fmt.Println("Error updating user role to admin:", err)
			return err
		}
		return nil
	}
}
func (r *roleRequestRepository) Reject(ctx context.Context, id uint, admin string) error {
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&entity.RoleRequest{}).
		Where("id = ? AND status = ?", id, entity.RoleReqPending).
		Updates(map[string]interface{}{"status": entity.RoleReqRejected, "decided_by": admin, "decided_at": &now}).Error
	if err != nil {
		fmt.Println("Error rejecting role request:", err)
		return err
	}
	return nil
}

func (r *roleRequestRepository) GetByID(ctx context.Context, id uint) (*entity.RoleRequest, error) {
	var rr entity.RoleRequest
	if err := r.db.WithContext(ctx).First(&rr, id).Error; err != nil {
		return nil, err
	}
	return &rr, nil
}
