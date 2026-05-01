package order

// FailureReason represents why an order has failed.
type FailureReason string

const (
	FailureReasonNone                  FailureReason = "none"
	FailureReasonInsufficientInventory FailureReason = "insufficient_inventory"
	FailureReasonPaymentFail           FailureReason = "payment_fail"
	FailureReasonPublishFail           FailureReason = "publish_fail"
)

func (r FailureReason) String() string {
	return string(r)
}
