package models

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
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	Price       float64 `json:"price" validate:"required,min=0"`
}

// PutItemRequest is the request body for the put item endpoint
type PutItemRequest Item

type Role int

const (
	RoleUser Role = iota + 1
	RoleManager
	RoleAdmin
)

type User struct {
	ID   string `json:"id"`
	Role Role   `json:"role"`
}

const (
	ActionCreate = "create"
	ActionDelete = "delete"
	ActionUpdate = "update"
)
