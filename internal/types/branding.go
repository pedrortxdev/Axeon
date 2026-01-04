package types

import "time"

type BrandingSettings struct {
	ID            int       `json:"id" db:"id"`
	UserID        int       `json:"user_id" db:"user_id"`
	LogoURL       string    `json:"logo_url" db:"logo_url"`
	PrimaryColor  string    `json:"primary_color" db:"primary_color"`
	HidePoweredBy bool      `json:"hide_powered_by" db:"hide_powered_by"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
