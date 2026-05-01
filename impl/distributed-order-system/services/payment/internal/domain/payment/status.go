package payment

// PaymentStatus represents the current state of a payment.
type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusSucceeded PaymentStatus = "succeeded"
	StatusFailed    PaymentStatus = "failed"
)

func (s PaymentStatus) String() string {
	return string(s)
}
