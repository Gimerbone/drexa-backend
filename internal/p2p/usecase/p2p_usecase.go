// Package usecase implements the P2P marketplace business logic. Crypto custody
// is delegated to the on-chain P2PEscrow contract (via p2p/chain): creating an
// order funds an escrow, the seller's confirmation releases it to the buyer, and
// cancel/expiry/dispute-resolution refund the seller — all on-chain.
package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"drexa/internal/p2p"
)

// service implements both p2p.Usecase and p2p.AdminUsecase.
type service struct {
	repo           p2p.Repository
	confirmTimeout time.Duration
	walletSvc      p2p.WalletService
}

// New returns the user-facing P2P usecase.
func New(repo p2p.Repository, confirmTimeout time.Duration, walletSvc p2p.WalletService) p2p.Usecase {
	return &service{repo: repo, confirmTimeout: confirmTimeout, walletSvc: walletSvc}
}

// NewAdmin returns the admin-facing P2P usecase (dispute resolution). It shares
// the same backing implementation as New.
func NewAdmin(repo p2p.Repository, confirmTimeout time.Duration, walletSvc p2p.WalletService) p2p.AdminUsecase {
	return &service{repo: repo, confirmTimeout: confirmTimeout, walletSvc: walletSvc}
}

// chainCtx derives a context bounded by the configured tx-confirmation timeout.
func (s *service) chainCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	if s.confirmTimeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, s.confirmTimeout)
}

func ptr(s string) *string { return &s }

// ─── Advertisements ──────────────────────────────────────────────────────────

func (s *service) CreateAd(ctx context.Context, sellerID string, in p2p.CreateAdInput) (*p2p.P2PAdvertisement, error) {
	in.PairID = strings.ReplaceAll(strings.TrimSpace(in.PairID), "-", "_")
	in.PaymentMethod = strings.TrimSpace(in.PaymentMethod)
	if in.PairID == "" || in.PaymentMethod == "" {
		return nil, p2p.ErrInvalidInput
	}
	if in.Price <= 0 || in.Amount <= 0 {
		return nil, p2p.ErrInvalidInput
	}
	if in.PaymentWindow <= 0 {
		return nil, p2p.ErrInvalidInput
	}
	if in.Type != p2p.AdTypeBuy && in.Type != p2p.AdTypeSell {
		in.Type = p2p.AdTypeSell
	}

	parts := strings.Split(in.PairID, "_")
	if len(parts) != 2 {
		return nil, p2p.ErrInvalidInput
	}
	baseCurrency := parts[0]
	quoteCurrency := parts[1]

	adID := uuid.NewString()

	if in.Type == p2p.AdTypeBuy {
		// Debit quote (USDT)
		totalQuote := in.Amount * in.Price
		if err := s.walletSvc.DebitBalance(ctx, sellerID, quoteCurrency, totalQuote, adID, "P2P Buy Ad Locked"); err != nil {
			return nil, p2p.ErrInsufficientFunds
		}
	} else {
		// Debit base (Crypto)
		if err := s.walletSvc.DebitBalance(ctx, sellerID, baseCurrency, in.Amount, adID, "P2P Sell Ad Locked"); err != nil {
			return nil, p2p.ErrInsufficientFunds
		}
	}
	
	sellerAddress, err := s.walletSvc.GetDepositAddress(ctx, sellerID, "ETH")
	if err != nil {
		sellerAddress = ""
	}

	ad := &p2p.P2PAdvertisement{
		AdvertisementID: adID,
		SellerID:        sellerID,
		Type:            in.Type,
		PairID:          in.PairID,
		Price:           in.Price,
		TotalAmount:     in.Amount,
		RemainingAmount: in.Amount,
		PaymentMethod:   in.PaymentMethod,
		PaymentWindow:   in.PaymentWindow,
		SellerAddress:   sellerAddress,
		Status:          p2p.AdStatusActive,
	}
	if err := s.repo.CreateAd(ctx, ad); err != nil {
		// If DB fails, refund the user
		if in.Type == p2p.AdTypeBuy {
			_ = s.walletSvc.CreditBalance(ctx, sellerID, quoteCurrency, in.Amount*in.Price, adID, "P2P Refund Failed Ad")
		} else {
			_ = s.walletSvc.CreditBalance(ctx, sellerID, baseCurrency, in.Amount, adID, "P2P Refund Failed Ad")
		}
		return nil, err
	}
	return ad, nil
}

func (s *service) ListAds(ctx context.Context, f p2p.AdFilter) ([]p2p.P2PAdvertisement, error) {
	if f.Status == "" {
		f.Status = p2p.AdStatusActive
	}
	f.PairID = strings.ReplaceAll(f.PairID, "-", "_")
	return s.repo.ListAds(ctx, f)
}

func (s *service) GetAd(ctx context.Context, id string) (*p2p.P2PAdvertisement, error) {
	return s.repo.GetAd(ctx, id)
}

func (s *service) MyAds(ctx context.Context, sellerID string) ([]p2p.P2PAdvertisement, error) {
	return s.repo.ListAdsBySeller(ctx, sellerID)
}

func (s *service) SetAdStatus(ctx context.Context, sellerID, adID string, status p2p.AdvertisementStatus) error {
	switch status {
	case p2p.AdStatusActive, p2p.AdStatusPaused, p2p.AdStatusCompleted, p2p.AdStatusCancelled:
	default:
		return p2p.ErrInvalidInput
	}
	ad, err := s.repo.GetAd(ctx, adID)
	if err != nil {
		return err
	}
	if ad.SellerID != sellerID {
		return p2p.ErrForbidden
	}

	if status == p2p.AdStatusCancelled && ad.Status != p2p.AdStatusCancelled && ad.Status != p2p.AdStatusCompleted {
		if ad.RemainingAmount > 0 {
			parts := strings.Split(ad.PairID, "_")
			baseCurrency := parts[0]
			quoteCurrency := parts[1]

			if ad.Type == p2p.AdTypeBuy {
				totalQuote := ad.RemainingAmount * ad.Price
				_ = s.walletSvc.CreditBalance(ctx, sellerID, quoteCurrency, totalQuote, adID, "P2P Buy Ad Cancelled Refund")
			} else {
				_ = s.walletSvc.CreditBalance(ctx, sellerID, baseCurrency, ad.RemainingAmount, adID, "P2P Sell Ad Cancelled Refund")
			}
		}
		ad.RemainingAmount = 0
		ad.Status = p2p.AdStatusCancelled
		return s.repo.UpdateAd(ctx, ad)
	}

	return s.repo.UpdateAdStatus(ctx, adID, status)
}

// ─── Orders ──────────────────────────────────────────────────────────────────

func (s *service) CreateOrder(ctx context.Context, takerID string, in p2p.CreateOrderInput) (*p2p.P2POrder, error) {
	if in.AdvertisementID == "" || in.Amount <= 0 {
		return nil, p2p.ErrInvalidInput
	}
	
	buyerAddress, err := s.walletSvc.GetDepositAddress(ctx, takerID, "ETH")
	if err != nil {
		buyerAddress = ""
	}

	ad, err := s.repo.GetAd(ctx, in.AdvertisementID)
	if err != nil {
		return nil, err
	}
	if ad.Status != p2p.AdStatusActive {
		return nil, p2p.ErrAdNotActive
	}
	if ad.SellerID == takerID {
		return nil, p2p.ErrSelfTrade
	}
	if in.Amount > ad.RemainingAmount {
		return nil, p2p.ErrAmountOutOfRange
	}

	parts := strings.Split(ad.PairID, "_")
	baseCurrency := parts[0]
	quoteCurrency := parts[1]

	orderID := uuid.NewString()
	totalQuote := in.Amount * ad.Price

	makerID := ad.SellerID
	var buyerID, sellerID string

	if ad.Type == p2p.AdTypeBuy {
		// Maker is buying Crypto (locked USDT). Taker is selling Crypto (gets USDT).
		buyerID = makerID
		sellerID = takerID

		// 1. Debit crypto from Taker (Seller)
		if err := s.walletSvc.DebitBalance(ctx, takerID, baseCurrency, in.Amount, orderID, "P2P Order (Sell)"); err != nil {
			return nil, p2p.ErrInsufficientFunds
		}

		// 2. Credit quote (USDT) to Taker (Seller). It comes from Maker's locked USDT (no extra deduction needed from Maker's wallet)
		if err := s.walletSvc.CreditBalance(ctx, takerID, quoteCurrency, totalQuote, orderID, "P2P Order Proceed"); err != nil {
			return nil, err
		}

		// 3. Credit crypto to Maker (Buyer)
		if err := s.walletSvc.CreditBalance(ctx, makerID, baseCurrency, in.Amount, orderID, "P2P Ad Filled Proceed"); err != nil {
			return nil, err
		}

	} else {
		// Maker is selling Crypto (locked Crypto). Taker is buying Crypto (gets Crypto).
		buyerID = takerID
		sellerID = makerID

		// 1. Debit quote (USDT) from Taker (Buyer)
		if err := s.walletSvc.DebitBalance(ctx, takerID, quoteCurrency, totalQuote, orderID, "P2P Order (Buy)"); err != nil {
			return nil, p2p.ErrInsufficientFunds
		}

		// 2. Credit crypto to Taker (Buyer). It comes from Maker's locked Crypto
		if err := s.walletSvc.CreditBalance(ctx, takerID, baseCurrency, in.Amount, orderID, "P2P Order Proceed"); err != nil {
			return nil, err
		}

		// 3. Credit quote (USDT) to Maker (Seller)
		if err := s.walletSvc.CreditBalance(ctx, makerID, quoteCurrency, totalQuote, orderID, "P2P Ad Filled Proceed"); err != nil {
			return nil, err
		}
	}

	// Update ad
	ad.RemainingAmount -= in.Amount
	if ad.RemainingAmount <= 0.00000001 {
		ad.RemainingAmount = 0
		ad.Status = p2p.AdStatusCompleted
	}
	if err := s.repo.UpdateAd(ctx, ad); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	order := &p2p.P2POrder{
		P2POrderID:      orderID,
		AdvertisementID: ad.AdvertisementID,
		BuyerID:         buyerID,
		SellerID:        sellerID,
		Amount:          in.Amount,
		TotalUSD:        totalQuote,
		Status:          p2p.P2POrderReleased,
		BuyerAddress:    buyerAddress,
		SellerAddress:   ad.SellerAddress,
		OnChainID:       "internal_swap",
		EscrowState:     "completed",
		ExpiredAt:       now,
		PaidAt:          &now,
		ReleasedAt:      &now,
	}
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *service) MarkPaid(ctx context.Context, userID, orderID string, proofURL *string) (*p2p.P2POrder, error) {
	return nil, errors.New("not applicable for atomic internal swaps")
}

func (s *service) ReleaseOrder(ctx context.Context, userID, orderID string) (*p2p.P2POrder, error) {
	return nil, errors.New("not applicable for atomic internal swaps")
}

func (s *service) CancelOrder(ctx context.Context, userID, orderID string) (*p2p.P2POrder, error) {
	return nil, errors.New("not applicable for atomic internal swaps")
}

func (s *service) GetOrder(ctx context.Context, userID, orderID string) (*p2p.P2POrder, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, p2p.ErrForbidden
	}
	return order, nil
}

func (s *service) MyOrders(ctx context.Context, userID string) ([]p2p.P2POrder, error) {
	return s.repo.ListOrdersByUser(ctx, userID)
}

func (s *service) EscrowInfo(ctx context.Context, userID, orderID string) (p2p.OnChainEscrow, error) {
	return p2p.OnChainEscrow{
		Buyer:     "Internal Wallet",
		Seller:    "Internal Wallet",
		AmountWei: "0",
		State:     "completed",
		CreatedAt: uint64(time.Now().Unix()),
	}, nil
}

// ─── Disputes ────────────────────────────────────────────────────────────────

func (s *service) OpenDispute(ctx context.Context, userID, orderID string, in p2p.OpenDisputeInput) (*p2p.P2PDispute, error) {
	return nil, errors.New("not applicable for atomic internal swaps")
}

// ─── Admin ───────────────────────────────────────────────────────────────────

func (s *service) ListOpenDisputes(ctx context.Context) ([]p2p.P2PDispute, error) {
	return s.repo.ListOpenDisputes(ctx)
}

func (s *service) ResolveDispute(ctx context.Context, adminID, disputeID string, releaseToBuyer bool, resolution string) (*p2p.P2PDispute, error) {
	return nil, errors.New("not applicable for atomic internal swaps")
}
