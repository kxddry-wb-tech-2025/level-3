package models

import (
	"encoding/json"
	"time"
)

// Item is the item model
type Item struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

// PostItemRequest is the request body for the post item endpoint
type PostItemRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description string  `json:"description" validate:"min=0,max=1000"`
	Quantity    int     `json:"quantity" validate:"required,min=0"`
	Price       float64 `json:"price" validate:"required,min=0"`
}

// PutItemRequest is the request body for the put item endpoint
type PutItemRequest Item

// Role is the role of the user.
type Role int

// Roles

const (
	// RoleUser is the user role.
	RoleUser Role = iota + 1
	// RoleManager is the manager role.
	RoleManager
	// RoleAdmin is the admin role.
	RoleAdmin
)

// User is the user model.
type User struct {
	ID   string `json:"id"`
	Role Role   `json:"role"`
}

// Actions

const (
	// ActionCreate is the create action.
	ActionCreate = "create"
	// ActionDelete is the delete action.
	ActionDelete = "delete"
	// ActionUpdate is the update action.
	ActionUpdate = "update"
)

// HistoryEntry is the history entry model.
type HistoryEntry struct {
	ID        string          `json:"id"`
	Action    string          `json:"action"`
	ItemID    string          `json:"item_id"`
	UserID    string          `json:"user_id"`
	UserRole  Role            `json:"user_role"`
	ChangedAt time.Time       `json:"changed_at"`
	OldData   json.RawMessage `json:"old_data"`
	NewData   json.RawMessage `json:"new_data"`
}
