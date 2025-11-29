package payments

import (
	"context"
	"time"

	"github.com/devchuckcamp/gocommerce/money"
)

// Gateway defines the payment gateway interface.
type Gateway interface {
	CreateIntent(ctx context.Context, req IntentRequest) (*PaymentIntent, error)
	GetIntent(ctx context.Context, intentID string) (*PaymentIntent, error)
	CaptureIntent(ctx context.Context, intentID string) (*PaymentIntent, error)
	CancelIntent(ctx context.Context, intentID string) (*PaymentIntent, error)
	CreateRefund(ctx context.Context, req RefundRequest) (*Refund, error)
	GetRefund(ctx context.Context, refundID string) (*Refund, error)
}

// PaymentIntent represents a payment intent.
type PaymentIntent struct {
	ID              string
	Amount          money.Money
	Currency        string
	Status          IntentStatus
	PaymentMethodID string
	OrderID         string
	Description     string
	CapturedAmount  money.Money
	Metadata        map[string]string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ExpiresAt       *time.Time
}

// IntentStatus represents the state of a payment intent.
type IntentStatus string

const (
	IntentStatusPending           IntentStatus = "pending"
	IntentStatusRequiresAction    IntentStatus = "requires_action"
	IntentStatusProcessing        IntentStatus = "processing"
	IntentStatusSucceeded         IntentStatus = "succeeded"
	IntentStatusCanceled          IntentStatus = "canceled"
	IntentStatusFailed            IntentStatus = "failed"
)

// IntentRequest contains data to create a payment intent.
type IntentRequest struct {
	Amount          money.Money
	Currency        string
	PaymentMethodID string
	OrderID         string
	Description     string
	Metadata        map[string]string
	CaptureMethod   CaptureMethod
}

// CaptureMethod defines when to capture payment.
type CaptureMethod string

const (
	CaptureMethodAutomatic CaptureMethod = "automatic"
	CaptureMethodManual    CaptureMethod = "manual"
)

// Refund represents a payment refund.
type Refund struct {
	ID              string
	PaymentIntentID string
	Amount          money.Money
	Currency        string
	Status          RefundStatus
	Reason          RefundReason
	Metadata        map[string]string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// RefundStatus represents the state of a refund.
type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusSucceeded RefundStatus = "succeeded"
	RefundStatusFailed    RefundStatus = "failed"
	RefundStatusCanceled  RefundStatus = "canceled"
)

// RefundReason represents why a refund was issued.
type RefundReason string

const (
	RefundReasonDuplicate          RefundReason = "duplicate"
	RefundReasonFraudulent         RefundReason = "fraudulent"
	RefundReasonRequestedByCustomer RefundReason = "requested_by_customer"
	RefundReasonDefectiveProduct   RefundReason = "defective_product"
	RefundReasonOther              RefundReason = "other"
)

// RefundRequest contains data to create a refund.
type RefundRequest struct {
	PaymentIntentID string
	Amount          money.Money
	Reason          RefundReason
	Metadata        map[string]string
}

// IsRefundable returns true if the intent can be refunded.
func (pi *PaymentIntent) IsRefundable() bool {
	return pi.Status == IntentStatusSucceeded &&
		!pi.CapturedAmount.IsZero()
}

// IsCancelable returns true if the intent can be canceled.
func (pi *PaymentIntent) IsCancelable() bool {
	return pi.Status == IntentStatusPending ||
		pi.Status == IntentStatusRequiresAction
}

// Repository defines methods for payment persistence.
type Repository interface {
	SaveIntent(ctx context.Context, intent *PaymentIntent) error
	FindIntent(ctx context.Context, intentID string) (*PaymentIntent, error)
	FindIntentsByOrder(ctx context.Context, orderID string) ([]*PaymentIntent, error)
	SaveRefund(ctx context.Context, refund *Refund) error
	FindRefund(ctx context.Context, refundID string) (*Refund, error)
	FindRefundsByIntent(ctx context.Context, intentID string) ([]*Refund, error)
}
