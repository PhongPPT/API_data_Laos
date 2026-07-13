package domain

import (
	"context"
)

type Province struct {
	ID     int      `json:"id" db:"pr_id"`
	Name   string   `json:"name" db:"pr_name"`
	NameEn string   `json:"name_en" db:"pr_name_en"`
	Lat    *float64 `json:"lat" db:"lat"`
	Lng    *float64 `json:"lng" db:"lng"`
}

type District struct {
	ID         int      `json:"id" db:"dr_id"`
	Name       string   `json:"name" db:"dr_name"`
	NameEn     string   `json:"name_en" db:"dr_name_en"`
	ProvinceID int      `json:"province_id" db:"pr_id"`
	Lat        *float64 `json:"lat" db:"lat"`
	Lng        *float64 `json:"lng" db:"lng"`
}

type Village struct {
	ID         int      `json:"id" db:"vill_id"`
	Name       string   `json:"name" db:"vill_name"`
	NameEn     string   `json:"name_en" db:"vill_name_en"`
	DistrictID int      `json:"district_id" db:"dr_id"`
	Lat        *float64 `json:"lat" db:"lat"`
	Lng        *float64 `json:"lng" db:"lng"`
}

// Pagination struct
type Pagination struct {
	CurrentPage          int `json:"current_page"`
	CurrentPageTotalItem int `json:"current_page_total_item"`
	TotalItems           int `json:"total_items"`
	TotalPage            int `json:"total_page"`
}

// ResponseData holds the list of items and pagination metadata
type ResponseData struct {
	ListData   interface{} `json:"list_data"`
	Pagination Pagination  `json:"pagination"`
}

// StandardResponse wraps the standard API response format
type StandardResponse struct {
	Status int           `json:"status"`
	Data   *ResponseData `json:"data"`
	Error  *string       `json:"error"`
}

// Usecase Interfaces
type GeoUsecase interface {
	GetProvinces(ctx context.Context, page, limit int) (StandardResponse, error)
	GetDistrictsByProvince(ctx context.Context, provinceID, page, limit int) (StandardResponse, error)
	GetVillagesByDistrict(ctx context.Context, districtID, page, limit int) (StandardResponse, error)
}

// Repository Interfaces
type GeoRepository interface {
	FetchProvinces(ctx context.Context, limit, offset int) ([]Province, int, error)
	FetchDistrictsByProvince(ctx context.Context, provinceID, limit, offset int) ([]District, int, error)
	FetchVillagesByDistrict(ctx context.Context, districtID, limit, offset int) ([]Village, int, error)
}
