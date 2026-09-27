package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	OwnerID   *int      `json:"owner_id,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateStudentRequest struct {
	NIM   string  `json:"nim" validate:"required,nim"`
	Name  string  `json:"name" validate:"required,min=3,max=100"`
	Grade float64 `json:"grade" validate:"min=0,max=100"`
}

type ReplaceStudentRequest struct {
	Name     string  `json:"name" validate:"required,min=3,max=100"`
	Grade    float64 `json:"grade" validate:"min=0,max=100"`
	IsActive bool    `json:"is_active"`
}

type PatchStudentRequest struct {
	Name     *string  `json:"name,omitempty" validate:"omitnil,min=3,max=100"`
	Grade    *float64 `json:"grade,omitempty" validate:"omitnil,min=0,max=100"`
	IsActive *bool    `json:"is_active,omitempty"`
}

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type Cursor struct {
	CreatedAt time.Time
	ID        int
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

type CursorQuery struct {
	Limit    int
	Search   string
	IsActive *bool
	After    *Cursor
}

type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
	MinGrade *float64
	MaxGrade *float64
}

func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
