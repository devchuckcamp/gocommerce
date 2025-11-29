package user

import (
	"context"
	"time"
)

// UserProfile represents a user's profile information.
type UserProfile struct {
	ID          string
	Email       string
	FirstName   string
	LastName    string
	Phone       string
	DateOfBirth *time.Time
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FullName returns the user's full name.
func (u *UserProfile) FullName() string {
	return u.FirstName + " " + u.LastName
}

// Address represents a user's saved address.
type Address struct {
	ID           string
	UserID       string
	Label        string // e.g., "Home", "Work"
	FirstName    string
	LastName     string
	Company      string
	AddressLine1 string
	AddressLine2 string
	City         string
	State        string
	PostalCode   string
	Country      string
	Phone        string
	IsDefault    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// FullName returns the full name from the address.
func (a *Address) FullName() string {
	return a.FirstName + " " + a.LastName
}

// IsComplete checks if the address has all required fields.
func (a *Address) IsComplete() bool {
	return a.FirstName != "" &&
		a.LastName != "" &&
		a.AddressLine1 != "" &&
		a.City != "" &&
		a.PostalCode != "" &&
		a.Country != ""
}

// ProfileRepository defines methods for user profile persistence.
type ProfileRepository interface {
	FindByID(ctx context.Context, id string) (*UserProfile, error)
	FindByEmail(ctx context.Context, email string) (*UserProfile, error)
	Save(ctx context.Context, profile *UserProfile) error
	Delete(ctx context.Context, id string) error
}

// AddressRepository defines methods for address persistence.
type AddressRepository interface {
	FindByID(ctx context.Context, id string) (*Address, error)
	FindByUserID(ctx context.Context, userID string) ([]*Address, error)
	FindDefaultByUserID(ctx context.Context, userID string) (*Address, error)
	Save(ctx context.Context, address *Address) error
	Delete(ctx context.Context, id string) error
	SetDefault(ctx context.Context, userID, addressID string) error
}
