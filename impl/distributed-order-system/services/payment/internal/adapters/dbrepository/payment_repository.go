package dbrepository

import (
	"context"
	"fmt"

	"payment-service/internal/domain/payment"

	"gorm.io/gorm"
)

// GormPaymentRepository implements payment.Repository using GORM + PostgreSQL.
type GormPaymentRepository struct {
	db *gorm.DB
}

// NewGormPaymentRepository creates a new repository instance.
func NewGormPaymentRepository(db *gorm.DB) *GormPaymentRepository {
	return &GormPaymentRepository{db: db}
}

// AutoMigrate runs GORM auto-migration for the persistence models.
func (r *GormPaymentRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&paymentModel{})
}

// Save persists a new payment (INSERT).
func (r *GormPaymentRepository) Save(ctx context.Context, p *payment.Payment) error {
	model := paymentToModel(p)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}
	return nil
}

// GetByID loads a payment by its ID.
func (r *GormPaymentRepository) GetByID(ctx context.Context, id string) (*payment.Payment, error) {
	var model paymentModel
	if err := r.db.WithContext(ctx).Where("payment_id = ?", id).First(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment %s: %w", id, err)
	}
	return modelToPayment(&model), nil
}

// Update persists status changes to an existing payment.
func (r *GormPaymentRepository) Update(ctx context.Context, p *payment.Payment) error {
	result := r.db.WithContext(ctx).Model(&paymentModel{}).Where("payment_id = ?", p.ID()).Updates(map[string]interface{}{
		"status":     p.Status().String(),
		"updated_at": p.UpdatedAt(),
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update payment %s: %w", p.ID(), result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("payment not found: %s", p.ID())
	}
	return nil
}

// paymentToModel converts a domain Payment to a GORM persistence model.
func paymentToModel(p *payment.Payment) *paymentModel {
	return &paymentModel{
		PaymentID:  p.ID(),
		OrderID:    p.OrderID(),
		TotalPrice: p.TotalPrice(),
		Status:     p.Status().String(),
		CreatedAt:  p.CreatedAt(),
		UpdatedAt:  p.UpdatedAt(),
	}
}

// modelToPayment converts a GORM persistence model to a domain Payment.
func modelToPayment(m *paymentModel) *payment.Payment {
	return payment.ReconstructPayment(
		m.PaymentID,
		m.OrderID,
		m.TotalPrice,
		payment.PaymentStatus(m.Status),
		m.CreatedAt,
		m.UpdatedAt,
	)
}
