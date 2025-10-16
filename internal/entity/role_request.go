package entity

import "time"

type RoleRequestStatus string

const (
	RoleReqPending  RoleRequestStatus = "pending"
	RoleReqApproved RoleRequestStatus = "approved"
	RoleReqRejected RoleRequestStatus = "rejected"
)

type RoleRequest struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	Username      string            `json:"username" gorm:"index"`
	RequestedRole string            `json:"requested_role"` // "admin"
	Status        RoleRequestStatus `json:"status" gorm:"default:pending"`
	Reason        string            `json:"reason"`     // opsiyonel (form ekleyebiliriz)
	DecidedBy     *string           `json:"decided_by"` // admin username
	DecidedAt     *time.Time        `json:"decided_at"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}
