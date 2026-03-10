package models

import "time"

type Asset struct {
	ID                 string     `db:"id" json:"id"`
	CurrentAssignedTo  *string    `db:"current_assigned_to" json:"current_assigned_to,omitempty"`
	Brand              string     `db:"brand" json:"brand"`
	Model              string     `db:"model" json:"model"`
	SerialNo           string     `db:"serial_no" json:"serial_no"`
	AssetType          string     `db:"asset_type" json:"asset_type"`
	Status             string     `db:"status" json:"status"`
	Owner              string     `db:"owner" json:"owner"`
	WarrantyStart      time.Time  `db:"warranty_start_date" json:"warranty_start_date"`
	WarrantyExpiryDate time.Time  `db:"warranty_expiry_date" json:"warranty_expiry_date"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	ArchivedAt         *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}

type AssetRequest struct {
	Brand              string    `json:"brand" validate:"required"`
	Model              string    `json:"model" validate:"required"`
	SerialNo           string    `json:"serial_no" validate:"required"`
	AssetType          string    `json:"asset_type" validate:"required,oneof=laptop mouse keyboard mobile"`
	Owner              string    `json:"owner" validate:"required,oneof=company client"`
	WarrantyStart      time.Time `json:"warranty_start_date"`
	WarrantyExpiryDate time.Time `json:"warranty_expiry_date"`

	Laptop   *LaptopRequest   `json:"laptop,omitempty"`
	Mobile   *MobileRequest   `json:"mobile,omitempty"`
	Keyboard *KeyboardRequest `json:"keyboard,omitempty"`
	Mouse    *MouseRequest    `json:"mouse,omitempty"`
}
type LaptopRequest struct {
	Processor string `json:"processor"`
	RAM       string `json:"ram"`
	Storage   string `json:"storage"`
	OS        string `json:"os"`
	Charger   string `json:"charger"`
	Password  string `json:"password" validate:"required"`
}

type KeyboardRequest struct {
	Layout       string `json:"layout"`
	Connectivity string `json:"connectivity"`
}
type MouseRequest struct {
	DPI          int    `json:"dpi"`
	Connectivity string `json:"connectivity"`
}
type AssignedAssetInfo struct {
	ID    string `db:"id" json:"-"`
	Brand string `db:"brand" json:"brand"`
	Model string `db:"model" json:"model"`
}
type MobileRequest struct {
	OS       string `json:"os" validate:"required"`
	RAM      string `json:"ram" validate:"required"`
	Storage  string `json:"storage" validate:"required"`
	Charger  string `json:"charger"`
	Password string `json:"password" validate:"required"`
}

type GetAllAssetsResponse struct {
	Summary AssetSummary `json:"summary"`
	Assets  []AssetList  `json:"assets"`
}
type AssetList struct {
	ID         string  `db:"id" json:"id"`
	Brand      string  `db:"brand" json:"brand"`
	Model      string  `db:"model" json:"model"`
	AssetType  string  `db:"asset_type" json:"asset_type"`
	SerialNo   string  `db:"serial_no" json:"serial_no"`
	Status     string  `db:"status" json:"status"`
	AssignedTo *string `db:"assigned_to" json:"assigned_to"`
	OwnedBy    string  `db:"owned_by" json:"owned_by"`
}
type AssetSummary struct {
	Total            int `db:"total" json:"total"`
	Available        int `db:"available" json:"available"`
	Assigned         int `db:"assigned" json:"assigned"`
	WaitingForRepair int `db:"waiting_for_repair" json:"waiting_for_repair"`
	Service          int `db:"service" json:"service"`
	Damaged          int `db:"damaged" json:"damaged"`
}

type UpdateAssetRequest struct {
	Brand              string    `json:"brand" validate:"required"`
	Model              string    `json:"model" validate:"required"`
	SerialNo           string    `json:"serial_no" validate:"required"`
	AssetType          string    `json:"asset_type" validate:"required,oneof=laptop mouse keyboard mobile"`
	Owner              string    `json:"owner" validate:"required,oneof=company client"`
	WarrantyStart      time.Time `json:"warranty_start_date" validate:"required"`
	WarrantyExpiryDate time.Time `json:"warranty_expiry_date" validate:"required"`

	Laptop   *LaptopRequest   `json:"laptop,omitempty"`
	Mouse    *MouseRequest    `json:"mouse,omitempty"`
	Keyboard *KeyboardRequest `json:"keyboard,omitempty"`
	Mobile   *MobileRequest   `json:"mobile,omitempty"`
}

type ServiceRequest struct {
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
}
