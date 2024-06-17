package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leetcode-golang-classroom/golang-order-api/internal/model"
	"github.com/leetcode-golang-classroom/golang-order-api/internal/repository/order"
	"github.com/leetcode-golang-classroom/golang-order-api/internal/util"
)

type Repo interface {
	Insert(ctx context.Context, order model.Order) error
	FindByID(ctx context.Context, id uint64) (model.Order, error)
	DeleteByID(ctx context.Context, id uint64) error
	Update(ctx context.Context, order model.Order) error
	FindAll(ctx context.Context, page order.FindAllPage) (order.FindResult, error)
}
type Order struct {
	Repo Repo
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		// w.WriteHeader(http.StatusBadRequest)
		util.WriteJSONError(w, http.StatusBadRequest, "payload incorrect")
		return
	}
	now := time.Now().UTC()
	order := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}
	err := o.Repo.Insert(r.Context(), order)
	if err != nil {
		log.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(order)
	if err != nil {
		log.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}
	const decimal = 10
	const bitSize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		util.WriteJSONError(w, http.StatusBadRequest, "cursor not in correct format")
		return
	}
	const size = 50
	res, err := o.Repo.FindAll(r.Context(), order.FindAllPage{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		log.Println("failed to find all:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var response struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}
	response.Items = res.Orders
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		log.Println("failed to marshal response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		util.WriteJSONError(w, http.StatusBadRequest, "id is not in correct format")
		return
	}
	orderRes, err := o.Repo.FindByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		util.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	} else if err != nil {
		log.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(orderRes); err != nil {
		log.Println("falied to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		util.WriteJSONError(w, http.StatusBadRequest, "status not provided")
		return
	}
	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	theOrder, err := o.Repo.FindByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		util.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		log.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	const completeStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now().UTC()
	switch body.Status {
	case shippedStatus:
		if theOrder.ShippedAt != nil {
			util.WriteJSONError(w, http.StatusBadRequest, "order has been shipped")
			return
		}
		theOrder.ShippedAt = &now
	case completeStatus:
		if theOrder.CompletedAt != nil || theOrder.ShippedAt == nil {
			util.WriteJSONError(w, http.StatusBadRequest, "order has not shipped or order has been completed")
			return
		}
		theOrder.CompletedAt = &now
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = o.Repo.Update(r.Context(), theOrder)
	if err != nil {
		log.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(&theOrder); err != nil {
		log.Println("failed to marshal", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		util.WriteJSONError(w, http.StatusBadRequest, "orderID not provided")
		return
	}
	err = o.Repo.DeleteByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		util.WriteJSONError(w, http.StatusNotFound, "order not found")
		return
	} else if err != nil {
		log.Println("failed to delete by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
