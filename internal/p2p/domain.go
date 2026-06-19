package p2p

import (
	"context"
	"errors"
	"time"
)

// ─── Enums ───────────────────────────────────────────────────────────────────

type AdvertisementStatus string

const (
	AdStatusActive    AdvertisementStatus = "active"
	AdStatusPaused    AdvertisementStatus = "paused"
	AdStatusCompleted AdvertisementStatus = "completed"
	AdStatusCancelled AdvertisementStatus = "cancelled"
)

type AdvertisementType string

const (
	AdTypeBuy  AdvertisementType = "buy"
	AdTypeSell AdvertisementType = "sell"
)

type OrderStatus string

const (
	P2POrderCreated  OrderStatus = "created"
	P2POrderPaid     OrderStatus = "paid"
	P2POrderReleased OrderStatus = "released"
	P2POrderDisputed OrderStatus = "disputed"
	P2POrderCancelled OrderStatus = "cancelled"
)

type DisputeStatus string

const (
	DisputeOpen     DisputeStatus = "open"
	DisputeResolved DisputeStatus = "resolved"
)

// ─── Entities ────────────────────────────────────────────────────────────────

// P2PAdvertisement is a seller's offer to sell crypto for IDR.
type P2PAdvertisement struct {
	AdvertisementID string              `json:"advertisement_id" gorm:"primaryKey;column:advertisement_id"`
	SellerID        string              `json:"seller_id" gorm:"column:seller_id;index"`
	PairID          string              `json:"pair_id" gorm:"column:pair_id"`        // FK to trading_pairs
	Type            AdvertisementType   `json:"type" gorm:"column:type;default:sell"`
	Price           float64             `json:"price" gorm:"column:price;type:numeric(36,18)"`
	TotalAmount     float64             `json:"total_amount" gorm:"column:total_amount;type:numeric(36,18)"`
	RemainingAmount float64             `json:"remaining_amount" gorm:"column:remaining_amount;type:numeric(36,18)"`
	PaymentMethod   string              `json:"payment_method" gorm:"column:payment_method"` // e.g. "BCA", "Mandiri"
	PaymentWindow   int                 `json:"payment_window" gorm:"column:payment_window"` // minutes buyer has to pay
	// SellerAddress is the seller's EVM payout address — escrow refunds (cancel,
	// expiry, dispute-for-seller) are sent here by the on-chain contract.
	SellerAddress string              `json:"seller_address" gorm:"column:seller_address"`
	Status        AdvertisementStatus `json:"status" gorm:"column:status;default:active"`
	CreatedAt     time.Time           `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (P2PAdvertisement) TableName() string {
	return "p2p_advertisements"
}

// P2POrder is a buyer's order against an advertisement.
// Crypto is held in an on-chain escrow contract until the seller confirms the
// buyer's fiat payment (release) or the order is cancelled/expired (refund).
type P2POrder struct {
	P2POrderID      string      `json:"p2p_order_id" gorm:"primaryKey;column:p2p_order_id"`
	AdvertisementID string      `json:"advertisement_id" gorm:"column:advertisement_id;index"`
	BuyerID         string      `json:"buyer_id" gorm:"column:buyer_id;index"`
	SellerID        string      `json:"seller_id" gorm:"column:seller_id;index"`
	Amount          float64     `json:"amount" gorm:"column:amount;type:numeric(36,18)"`    // crypto amount
	TotalUSD        float64     `json:"total_usd" gorm:"column:total_usd;type:numeric(20,2)"`  // fiat equivalent
	Status          OrderStatus `json:"status" gorm:"column:status;default:created"`
	PaymentProofURL *string     `json:"payment_proof_url" gorm:"column:payment_proof_url"`
	// EscrowWalletID is the legacy internal-ledger escrow wallet (nullable now
	// that escrow is on-chain). Kept for backward compatibility.
	EscrowWalletID *string `json:"escrow_wallet_id" gorm:"column:escrow_wallet_id"`

	// ── On-chain escrow fields ───────────────────────────────────────────────
	BuyerAddress  string `json:"buyer_address" gorm:"column:buyer_address"`  // EVM release destination
	SellerAddress string `json:"seller_address" gorm:"column:seller_address"` // EVM refund destination
	OnChainID     string `json:"on_chain_id" gorm:"column:on_chain_id"`    // bytes32 escrow key (hex)
	EscrowState   string `json:"escrow_state" gorm:"column:escrow_state;default:none"`
	CreateTxHash  *string `json:"create_tx_hash" gorm:"column:create_tx_hash"`
	ReleaseTxHash *string `json:"release_tx_hash" gorm:"column:release_tx_hash"`
	RefundTxHash  *string `json:"refund_tx_hash" gorm:"column:refund_tx_hash"`
	DisputeTxHash *string `json:"dispute_tx_hash" gorm:"column:dispute_tx_hash"`

	CreatedAt  time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	PaidAt     *time.Time `json:"paid_at" gorm:"column:paid_at"`
	ReleasedAt *time.Time `json:"released_at" gorm:"column:released_at"`
	ExpiredAt  time.Time  `json:"expired_at" gorm:"column:expired_at"`
}

func (P2POrder) TableName() string {
	return "p2p_orders"
}

// P2PDispute is raised when a buyer/seller cannot resolve a P2POrder themselves.
type P2PDispute struct {
	P2PDisputeID string        `json:"p2p_dispute_id" gorm:"primaryKey;column:p2p_dispute_id"`
	P2POrderID   string        `json:"p2p_order_id" gorm:"column:p2p_order_id;index"`
	RaisedBy     string        `json:"raised_by" gorm:"column:raised_by"` // user ID
	Reason       string        `json:"reason" gorm:"column:reason"`
	EvidenceURL  *string       `json:"evidence_url" gorm:"column:evidence_url"`
	Status       DisputeStatus `json:"status" gorm:"column:status;default:open"`
	ResolvedBy   *string       `json:"resolved_by" gorm:"column:resolved_by"` // admin user ID
	Resolution   string        `json:"resolution" gorm:"column:resolution;default:''"`
	ResolvedAt   *time.Time    `json:"resolved_at" gorm:"column:resolved_at"`
	CreatedAt    time.Time     `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (P2PDispute) TableName() string {
	return "p2p_disputes"
}

// ─── Domain Errors ───────────────────────────────────────────────────────────

var (
	ErrAdvertisementNotFound = errors.New("advertisement not found")
	ErrAdNotActive           = errors.New("advertisement is not active")
	ErrP2POrderNotFound      = errors.New("p2p order not found")
	ErrP2POrderExpired       = errors.New("p2p order has expired")
	ErrDisputeNotFound       = errors.New("dispute not found")
	ErrDisputeExists         = errors.New("a dispute is already open for this order")
	ErrEscrowNotReleased     = errors.New("escrow has not been released")

	ErrAmountOutOfRange = errors.New("amount is outside the advertisement's min/max range")
	ErrSelfTrade        = errors.New("cannot create an order against your own advertisement")
	ErrForbidden        = errors.New("not a participant in this order")
	ErrInvalidState     = errors.New("order is not in a valid state for this action")
	ErrInvalidAddress   = errors.New("a valid EVM payout address is required")
	ErrInvalidInput     = errors.New("invalid input")
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrCryptoAddressNotFound = errors.New("crypto address not generated for user")
)

// ─── Filters & DTOs ────────────────────────────────────────────────────────

// AdFilter narrows an advertisement listing query. Empty fields are ignored.
type AdFilter struct {
	Type          AdvertisementType
	PairID        string
	PaymentMethod string
	Status        AdvertisementStatus
	Limit         int
	Offset        int
}

// CreateAdInput is the payload to post a sell advertisement.
type CreateAdInput struct {
	Type          AdvertisementType
	PairID        string
	Price         float64
	Amount        float64
	PaymentMethod string
	PaymentWindow int    // minutes
}

// CreateOrderInput is the payload to open an order against an advertisement.
type CreateOrderInput struct {
	AdvertisementID string
	Amount          float64 // crypto amount to buy
}

// OpenDisputeInput is the payload to raise a dispute on an order.
type OpenDisputeInput struct {
	Reason      string
	EvidenceURL *string
}

// OnChainEscrow is a chain-agnostic view of an escrow's on-chain state.
type OnChainEscrow struct {
	Buyer     string `json:"buyer"`
	Seller    string `json:"seller"`
	AmountWei string `json:"amount_wei"`
	State     string `json:"state"`
	CreatedAt uint64 `json:"created_at"`
}

// ─── Ports (interfaces) ──────────────────────────────────────────────────────

// WalletService integrates the P2P engine with the user's internal asset ledger
// and their on-chain deposit addresses.
type WalletService interface {
	// GetDepositAddress returns the user's deposit address for the given chain/currency.
	GetDepositAddress(ctx context.Context, userID, currency string) (string, error)
	// DebitBalance deducts an amount from the user's internal ledger.
	DebitBalance(ctx context.Context, userID, currency string, amount float64, refID, description string) error
	// CreditBalance adds an amount to the user's internal ledger.
	CreditBalance(ctx context.Context, userID, currency string, amount float64, refID, description string) error
}

// Repository persists P2P advertisements, orders, and disputes.
type Repository interface {
	// Advertisements
	CreateAd(ctx context.Context, ad *P2PAdvertisement) error
	GetAd(ctx context.Context, id string) (*P2PAdvertisement, error)
	ListAds(ctx context.Context, f AdFilter) ([]P2PAdvertisement, error)
	ListAdsBySeller(ctx context.Context, sellerID string) ([]P2PAdvertisement, error)
	UpdateAd(ctx context.Context, ad *P2PAdvertisement) error
	UpdateAdStatus(ctx context.Context, id string, status AdvertisementStatus) error

	// Orders
	CreateOrder(ctx context.Context, o *P2POrder) error
	GetOrder(ctx context.Context, id string) (*P2POrder, error)
	UpdateOrder(ctx context.Context, o *P2POrder) error
	ListOrdersByUser(ctx context.Context, userID string) ([]P2POrder, error)

	// Disputes
	CreateDispute(ctx context.Context, d *P2PDispute) error
	GetDispute(ctx context.Context, id string) (*P2PDispute, error)
	GetDisputeByOrder(ctx context.Context, orderID string) (*P2PDispute, error)
	ListOpenDisputes(ctx context.Context) ([]P2PDispute, error)
	UpdateDispute(ctx context.Context, d *P2PDispute) error
}

// Usecase is the user-facing P2P marketplace API (advertisements, orders,
// disputes), backed by an on-chain escrow contract.
type Usecase interface {
	// Advertisements
	CreateAd(ctx context.Context, sellerID string, in CreateAdInput) (*P2PAdvertisement, error)
	ListAds(ctx context.Context, f AdFilter) ([]P2PAdvertisement, error)
	GetAd(ctx context.Context, id string) (*P2PAdvertisement, error)
	MyAds(ctx context.Context, sellerID string) ([]P2PAdvertisement, error)
	SetAdStatus(ctx context.Context, sellerID, adID string, status AdvertisementStatus) error

	// Orders
	CreateOrder(ctx context.Context, buyerID string, in CreateOrderInput) (*P2POrder, error)
	MarkPaid(ctx context.Context, userID, orderID string, proofURL *string) (*P2POrder, error)
	ReleaseOrder(ctx context.Context, userID, orderID string) (*P2POrder, error)
	CancelOrder(ctx context.Context, userID, orderID string) (*P2POrder, error)
	GetOrder(ctx context.Context, userID, orderID string) (*P2POrder, error)
	MyOrders(ctx context.Context, userID string) ([]P2POrder, error)
	EscrowInfo(ctx context.Context, userID, orderID string) (OnChainEscrow, error)

	// Disputes
	OpenDispute(ctx context.Context, userID, orderID string, in OpenDisputeInput) (*P2PDispute, error)
}

// AdminUsecase is the admin-facing P2P dispute resolution API.
type AdminUsecase interface {
	ListOpenDisputes(ctx context.Context) ([]P2PDispute, error)
	ResolveDispute(ctx context.Context, adminID, disputeID string, releaseToBuyer bool, resolution string) (*P2PDispute, error)
}
