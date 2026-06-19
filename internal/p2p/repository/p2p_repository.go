package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"drexa/internal/p2p"
)

type p2pRepository struct{ db *gorm.DB }

// New returns a GORM-backed p2p.Repository.
func New(db *gorm.DB) p2p.Repository {
	return &p2pRepository{db: db}
}

// ─── Advertisements ──────────────────────────────────────────────────────────

func (r *p2pRepository) CreateAd(ctx context.Context, ad *p2p.P2PAdvertisement) error {
	if err := r.db.WithContext(ctx).Create(ad).Error; err != nil {
		return fmt.Errorf("p2p_repo: create ad: %w", err)
	}
	return nil
}

func (r *p2pRepository) GetAd(ctx context.Context, id string) (*p2p.P2PAdvertisement, error) {
	var ad p2p.P2PAdvertisement
	err := r.db.WithContext(ctx).Where("advertisement_id = ?", id).First(&ad).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, p2p.ErrAdvertisementNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("p2p_repo: get ad: %w", err)
	}
	return &ad, nil
}

func (r *p2pRepository) ListAds(ctx context.Context, f p2p.AdFilter) ([]p2p.P2PAdvertisement, error) {
	q := r.db.WithContext(ctx).Model(&p2p.P2PAdvertisement{})
	if f.PairID != "" {
		q = q.Where("pair_id = ?", f.PairID)
	}
	if f.PaymentMethod != "" {
		q = q.Where("payment_method = ?", f.PaymentMethod)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}

	var ads []p2p.P2PAdvertisement
	if err := q.Order("created_at DESC").Find(&ads).Error; err != nil {
		return nil, fmt.Errorf("p2p_repo: list ads: %w", err)
	}
	return ads, nil
}

func (r *p2pRepository) ListAdsBySeller(ctx context.Context, sellerID string) ([]p2p.P2PAdvertisement, error) {
	var ads []p2p.P2PAdvertisement
	if err := r.db.WithContext(ctx).
		Where("seller_id = ?", sellerID).
		Order("created_at DESC").
		Find(&ads).Error; err != nil {
		return nil, fmt.Errorf("p2p_repo: list ads by seller: %w", err)
	}
	return ads, nil
}

func (r *p2pRepository) UpdateAd(ctx context.Context, ad *p2p.P2PAdvertisement) error {
	if err := r.db.WithContext(ctx).Save(ad).Error; err != nil {
		return fmt.Errorf("p2p_repo: update ad: %w", err)
	}
	return nil
}

func (r *p2pRepository) UpdateAdStatus(ctx context.Context, id string, status p2p.AdvertisementStatus) error {
	res := r.db.WithContext(ctx).Model(&p2p.P2PAdvertisement{}).
		Where("advertisement_id = ?", id).
		Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("p2p_repo: update ad status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return p2p.ErrAdvertisementNotFound
	}
	return nil
}

// ─── Orders ──────────────────────────────────────────────────────────────────

func (r *p2pRepository) CreateOrder(ctx context.Context, o *p2p.P2POrder) error {
	if err := r.db.WithContext(ctx).Create(o).Error; err != nil {
		return fmt.Errorf("p2p_repo: create order: %w", err)
	}
	return nil
}

func (r *p2pRepository) GetOrder(ctx context.Context, id string) (*p2p.P2POrder, error) {
	var o p2p.P2POrder
	err := r.db.WithContext(ctx).Where("p2p_order_id = ?", id).First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, p2p.ErrP2POrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("p2p_repo: get order: %w", err)
	}
	return &o, nil
}

func (r *p2pRepository) UpdateOrder(ctx context.Context, o *p2p.P2POrder) error {
	if err := r.db.WithContext(ctx).Save(o).Error; err != nil {
		return fmt.Errorf("p2p_repo: update order: %w", err)
	}
	return nil
}

func (r *p2pRepository) ListOrdersByUser(ctx context.Context, userID string) ([]p2p.P2POrder, error) {
	var orders []p2p.P2POrder
	if err := r.db.WithContext(ctx).
		Where("buyer_id = ? OR seller_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("p2p_repo: list orders by user: %w", err)
	}
	return orders, nil
}

// ─── Disputes ────────────────────────────────────────────────────────────────

func (r *p2pRepository) CreateDispute(ctx context.Context, d *p2p.P2PDispute) error {
	if err := r.db.WithContext(ctx).Create(d).Error; err != nil {
		return fmt.Errorf("p2p_repo: create dispute: %w", err)
	}
	return nil
}

func (r *p2pRepository) GetDispute(ctx context.Context, id string) (*p2p.P2PDispute, error) {
	var d p2p.P2PDispute
	err := r.db.WithContext(ctx).Where("p2p_dispute_id = ?", id).First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, p2p.ErrDisputeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("p2p_repo: get dispute: %w", err)
	}
	return &d, nil
}

func (r *p2pRepository) GetDisputeByOrder(ctx context.Context, orderID string) (*p2p.P2PDispute, error) {
	var d p2p.P2PDispute
	err := r.db.WithContext(ctx).
		Where("p2p_order_id = ?", orderID).
		Order("created_at DESC").
		First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, p2p.ErrDisputeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("p2p_repo: get dispute by order: %w", err)
	}
	return &d, nil
}

func (r *p2pRepository) ListOpenDisputes(ctx context.Context) ([]p2p.P2PDispute, error) {
	var disputes []p2p.P2PDispute
	if err := r.db.WithContext(ctx).
		Where("status = ?", p2p.DisputeOpen).
		Order("created_at ASC").
		Find(&disputes).Error; err != nil {
		return nil, fmt.Errorf("p2p_repo: list open disputes: %w", err)
	}
	return disputes, nil
}

func (r *p2pRepository) UpdateDispute(ctx context.Context, d *p2p.P2PDispute) error {
	if err := r.db.WithContext(ctx).Save(d).Error; err != nil {
		return fmt.Errorf("p2p_repo: update dispute: %w", err)
	}
	return nil
}
