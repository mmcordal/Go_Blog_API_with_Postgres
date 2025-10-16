package viewmodel

import (
	"cleanArch_with_postgres/internal/entity"
	"time"
)

type RoleRequestVM struct {
	ID            uint       `json:"id"`
	Username      string     `json:"username"`
	RequestedRole string     `json:"requested_role"`
	Status        string     `json:"status"`
	Reason        string     `json:"reason"`
	DecidedBy     *string    `json:"decided_by"`
	DecidedAt     *time.Time `json:"decided_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

func ToRoleReqVM(r *entity.RoleRequest) *RoleRequestVM {
	return &RoleRequestVM{
		ID: r.ID, Username: r.Username, RequestedRole: r.RequestedRole,
		Status: string(r.Status), Reason: r.Reason,
		DecidedBy: r.DecidedBy, DecidedAt: r.DecidedAt, CreatedAt: r.CreatedAt,
	}
}

func ToRoleReqVMs(rs []entity.RoleRequest) []RoleRequestVM {
	out := make([]RoleRequestVM, len(rs))
	for i := range rs {
		out[i] = *ToRoleReqVM(&rs[i])
	}
	return out
}
