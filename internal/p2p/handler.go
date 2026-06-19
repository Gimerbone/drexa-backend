package p2p

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"

	"drexa/internal/p2p/chain"
)

// Handler exposes the P2P marketplace over HTTP. getUserID extracts the caller's
// user ID from the request (populated by the JWT middleware).
type Handler struct {
	uc        Usecase
	adminUc   AdminUsecase
	getUserID func(r *http.Request) string
}

func NewHandler(uc Usecase, adminUc AdminUsecase, getUserID func(r *http.Request) string) *Handler {
	return &Handler{uc: uc, adminUc: adminUc, getUserID: getUserID}
}

type messageResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func sendJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeErr maps domain errors to HTTP status codes.
func writeErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrAdvertisementNotFound),
		errors.Is(err, ErrP2POrderNotFound),
		errors.Is(err, ErrDisputeNotFound):
		sendJSON(w, http.StatusNotFound, messageResponse{Error: err.Error()})
	case errors.Is(err, ErrForbidden):
		sendJSON(w, http.StatusForbidden, messageResponse{Error: err.Error()})
	case errors.Is(err, ErrInvalidInput),
		errors.Is(err, ErrInvalidAddress),
		errors.Is(err, ErrAmountOutOfRange),
		errors.Is(err, ErrSelfTrade),
		errors.Is(err, ErrInsufficientFunds),
		errors.Is(err, ErrCryptoAddressNotFound):
		sendJSON(w, http.StatusBadRequest, messageResponse{Error: err.Error()})
	case errors.Is(err, ErrAdNotActive),
		errors.Is(err, ErrInvalidState),
		errors.Is(err, ErrP2POrderExpired),
		errors.Is(err, ErrDisputeExists):
		sendJSON(w, http.StatusConflict, messageResponse{Error: err.Error()})
	case errors.Is(err, chain.ErrNotConfigured):
		sendJSON(w, http.StatusServiceUnavailable, messageResponse{
			Error: "on-chain escrow is not configured on this server",
		})
	default:
		log.Error().Err(err).Msg("p2p: request failed")
		sendJSON(w, http.StatusInternalServerError, messageResponse{Error: "internal server error"})
	}
}

func (h *Handler) caller(w http.ResponseWriter, r *http.Request) (string, bool) {
	uid := h.getUserID(r)
	if uid == "" {
		sendJSON(w, http.StatusUnauthorized, messageResponse{Error: "unauthorized"})
		return "", false
	}
	return uid, true
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		sendJSON(w, http.StatusBadRequest, messageResponse{Error: "invalid JSON body"})
		return false
	}
	return true
}

// ─── Advertisement handlers ──────────────────────────────────────────────────

type createAdRequest struct {
	Type          AdvertisementType `json:"type"`
	PairID        string                `json:"pair_id"`
	Price         float64               `json:"price"`
	Amount        float64               `json:"amount"`
	PaymentMethod string                `json:"payment_method"`
	PaymentWindow int                   `json:"payment_window"`
}

// POST /api/v1/p2p/ads
func (h *Handler) CreateAd(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	var req createAdRequest
	if !decode(w, r, &req) {
		return
	}
	ad, err := h.uc.CreateAd(r.Context(), uid, CreateAdInput{
		Type:          req.Type,
		PairID:        req.PairID,
		Price:         req.Price,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		PaymentWindow: req.PaymentWindow,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusCreated, ad)
}

// GET /api/v1/p2p/ads
func (h *Handler) ListAds(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	ads, err := h.uc.ListAds(r.Context(), AdFilter{
		PairID:        q.Get("pair_id"),
		PaymentMethod: q.Get("payment_method"),
		Status:        AdvertisementStatus(q.Get("status")),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, ads)
}

// GET /api/v1/p2p/ads/mine
func (h *Handler) MyAds(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	ads, err := h.uc.MyAds(r.Context(), uid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, ads)
}

// GET /api/v1/p2p/ads/{id}
func (h *Handler) GetAd(w http.ResponseWriter, r *http.Request) {
	ad, err := h.uc.GetAd(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, ad)
}

type setAdStatusRequest struct {
	Status string `json:"status"`
}

// POST /api/v1/p2p/ads/{id}/status
func (h *Handler) SetAdStatus(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	var req setAdStatusRequest
	if !decode(w, r, &req) {
		return
	}
	if err := h.uc.SetAdStatus(r.Context(), uid, r.PathValue("id"), AdvertisementStatus(req.Status)); err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, messageResponse{Message: "advertisement status updated"})
}

// ─── Order handlers ──────────────────────────────────────────────────────────

type createOrderRequest struct {
	AdvertisementID string  `json:"advertisement_id"`
	Amount          float64 `json:"amount"`
}

// POST /api/v1/p2p/orders
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	var req createOrderRequest
	if !decode(w, r, &req) {
		return
	}
	order, err := h.uc.CreateOrder(r.Context(), uid, CreateOrderInput{
		AdvertisementID: req.AdvertisementID,
		Amount:          req.Amount,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusCreated, order)
}

// GET /api/v1/p2p/orders/mine
func (h *Handler) MyOrders(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	orders, err := h.uc.MyOrders(r.Context(), uid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, orders)
}

// GET /api/v1/p2p/orders/{id}
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	order, err := h.uc.GetOrder(r.Context(), uid, r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, order)
}

// GET /api/v1/p2p/orders/{id}/escrow
func (h *Handler) EscrowInfo(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	info, err := h.uc.EscrowInfo(r.Context(), uid, r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, info)
}

type markPaidRequest struct {
	PaymentProofURL *string `json:"payment_proof_url"`
}

// POST /api/v1/p2p/orders/{id}/paid
func (h *Handler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	var req markPaidRequest
	// Body is optional (proof URL); ignore decode errors on empty bodies.
	_ = json.NewDecoder(r.Body).Decode(&req)
	order, err := h.uc.MarkPaid(r.Context(), uid, r.PathValue("id"), req.PaymentProofURL)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, order)
}

// POST /api/v1/p2p/orders/{id}/release
func (h *Handler) ReleaseOrder(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	order, err := h.uc.ReleaseOrder(r.Context(), uid, r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, order)
}

// POST /api/v1/p2p/orders/{id}/cancel
func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	order, err := h.uc.CancelOrder(r.Context(), uid, r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, order)
}

type openDisputeRequest struct {
	Reason      string  `json:"reason"`
	EvidenceURL *string `json:"evidence_url"`
}

// POST /api/v1/p2p/orders/{id}/dispute
func (h *Handler) OpenDispute(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	var req openDisputeRequest
	if !decode(w, r, &req) {
		return
	}
	dispute, err := h.uc.OpenDispute(r.Context(), uid, r.PathValue("id"), OpenDisputeInput{
		Reason:      req.Reason,
		EvidenceURL: req.EvidenceURL,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusCreated, dispute)
}

// ─── Admin handlers ──────────────────────────────────────────────────────────

// GET /api/v1/admin/p2p/disputes
func (h *Handler) AdminListDisputes(w http.ResponseWriter, r *http.Request) {
	disputes, err := h.adminUc.ListOpenDisputes(r.Context())
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, disputes)
}

type resolveDisputeRequest struct {
	ReleaseToBuyer bool   `json:"release_to_buyer"`
	Resolution     string `json:"resolution"`
}

// POST /api/v1/admin/p2p/disputes/{id}/resolve
func (h *Handler) AdminResolveDispute(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.caller(w, r)
	if !ok {
		return
	}
	var req resolveDisputeRequest
	if !decode(w, r, &req) {
		return
	}
	dispute, err := h.adminUc.ResolveDispute(r.Context(), uid, r.PathValue("id"), req.ReleaseToBuyer, req.Resolution)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sendJSON(w, http.StatusOK, dispute)
}
