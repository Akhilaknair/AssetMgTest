package models

import (
	"time"

	"github.com/google/uuid"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginData struct {
	ID           string `db:"id"`
	PasswordHash string `db:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	PhoneNo  string `json:"phone_no" validate:"required"`
	//	Role        string    `json:"role" validate:"required,oneof=admin employee asset-manager project-manager"`
	UserType    string    `json:"user_type" validate:"required,oneof=full-time intern freelancer"`
	JoiningDate time.Time `json:"joining_date" db:"joining_date" validate:"required"`
}

type User struct {
	ID            string    `json:"userId" db:"id"`
	Name          string    `json:"name" db:"name"`
	Email         string    `json:"email" db:"email"`
	Role          string    `json:"role" db:"role"`
	PhoneNo       string    `json:"phoneNo" db:"phone_no"`
	UserType      string    `json:"userType" db:"user_type"`
	JoiningDate   time.Time `json:"joining_date" db:"joining_date"`
	AssignedCount int       `json:"assigned_count,omitempty"`
}

type AssignAssetRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" validate:"required"`
}

type UserWithAssets struct {
	ID             string              `db:"id" json:"userId"`
	Name           string              `db:"name" json:"name"`
	Email          string              `db:"email" json:"email"`
	PhoneNo        string              `db:"phone_no" json:"phoneNo"`
	UserType       string              `db:"user_type" json:"userType"`
	Role           string              `db:"role" json:"role"`
	AssignedCount  int                 `db:"assigned_count" json:"assigned_count"`
	AssignedAssets []AssignedAssetInfo `json:"assigned_assets"`
}
type Role string

const (
	Admin        Role = "admin"
	AssetManager Role = "asset-manager"
)

type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin employee asset-manager project-manager"`
}
