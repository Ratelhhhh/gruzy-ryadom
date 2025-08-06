package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"gruzy-ryadom/internal/models"
	"gruzy-ryadom/internal/service"
)

type API struct {
	service *service.Service
}

func New(service *service.Service) *API {
	return &API{service: service}
}

func (api *API) Routes() chi.Router {
	r := chi.NewRouter()
	
	// Public API
	r.Get("/v1/orders", api.GetOrders)
	
	return r
}

func (api *API) GetOrders(w http.ResponseWriter, r *http.Request) {
	filter := models.OrderFilter{}

	// Parse query parameters
	if minWeight := r.URL.Query().Get("min_weight"); minWeight != "" {
		if val, err := strconv.ParseFloat(minWeight, 64); err == nil {
			filter.MinWeight = val
		}
	}
	if maxWeight := r.URL.Query().Get("max_weight"); maxWeight != "" {
		if val, err := strconv.ParseFloat(maxWeight, 64); err == nil {
			filter.MaxWeight = val
		}
	}
	if minLength := r.URL.Query().Get("min_length"); minLength != "" {
		if val, err := strconv.ParseFloat(minLength, 64); err == nil {
			filter.MinLength = val
		}
	}
	if maxLength := r.URL.Query().Get("max_length"); maxLength != "" {
		if val, err := strconv.ParseFloat(maxLength, 64); err == nil {
			filter.MaxLength = val
		}
	}
	if minWidth := r.URL.Query().Get("min_width"); minWidth != "" {
		if val, err := strconv.ParseFloat(minWidth, 64); err == nil {
			filter.MinWidth = val
		}
	}
	if maxWidth := r.URL.Query().Get("max_width"); maxWidth != "" {
		if val, err := strconv.ParseFloat(maxWidth, 64); err == nil {
			filter.MaxWidth = val
		}
	}
	if minHeight := r.URL.Query().Get("min_height"); minHeight != "" {
		if val, err := strconv.ParseFloat(minHeight, 64); err == nil {
			filter.MinHeight = val
		}
	}
	if maxHeight := r.URL.Query().Get("max_height"); maxHeight != "" {
		if val, err := strconv.ParseFloat(maxHeight, 64); err == nil {
			filter.MaxHeight = val
		}
	}
	if minPrice := r.URL.Query().Get("min_price"); minPrice != "" {
		if val, err := strconv.ParseFloat(minPrice, 64); err == nil {
			filter.MinPrice = val
		}
	}
	if maxPrice := r.URL.Query().Get("max_price"); maxPrice != "" {
		if val, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			filter.MaxPrice = val
		}
	}
	if tags := r.URL.Query().Get("tags"); tags != "" {
		filter.Tags = strings.Split(tags, ",")
	}
	if from := r.URL.Query().Get("from"); from != "" {
		filter.From = from
	}
	if to := r.URL.Query().Get("to"); to != "" {
		filter.To = to
	}
	if page := r.URL.Query().Get("page"); page != "" {
		if val, err := strconv.Atoi(page); err == nil && val > 0 {
			filter.Page = val
		}
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil && val > 0 {
			filter.Limit = val
		}
	}
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}
	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	// Set defaults
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	orders, total, err := api.service.ListOrders(r.Context(), filter)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := models.OrdersResponse{
		Page:   filter.Page,
		Limit:  filter.Limit,
		Total:  total,
		Orders: orders,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
