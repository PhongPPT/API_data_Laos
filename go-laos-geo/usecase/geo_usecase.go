package usecase

import (
	"context"
	"math"

	"github.com/user/go-laos-geo/domain"
)

type geoUsecase struct {
	geoRepo domain.GeoRepository
}

func NewGeoUsecase(repo domain.GeoRepository) domain.GeoUsecase {
	return &geoUsecase{
		geoRepo: repo,
	}
}

func buildResponse(data interface{}, currentItemCount, total, page, limit int) domain.StandardResponse {
	totalPage := int(math.Ceil(float64(total) / float64(limit)))
	if totalPage == 0 {
		totalPage = 1
	}

	return domain.StandardResponse{
		Status: 1,
		Data: &domain.ResponseData{
			ListData: data,
			Pagination: domain.Pagination{
				CurrentPage:          page,
				CurrentPageTotalItem: currentItemCount,
				TotalItems:           total,
				TotalPage:            totalPage,
			},
		},
		Error: nil,
	}
}

func buildErrorResponse(err error) domain.StandardResponse {
	errStr := err.Error()
	return domain.StandardResponse{
		Status: 0,
		Data:   nil,
		Error:  &errStr,
	}
}

func calculateOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * limit
}

func (u *geoUsecase) GetProvinces(ctx context.Context, page, limit int) (domain.StandardResponse, error) {
	offset := calculateOffset(page, limit)
	provinces, total, err := u.geoRepo.FetchProvinces(ctx, limit, offset)
	if err != nil {
		return buildErrorResponse(err), err
	}
	return buildResponse(provinces, len(provinces), total, page, limit), nil
}

func (u *geoUsecase) GetDistrictsByProvince(ctx context.Context, provinceID, page, limit int) (domain.StandardResponse, error) {
	offset := calculateOffset(page, limit)
	districts, total, err := u.geoRepo.FetchDistrictsByProvince(ctx, provinceID, limit, offset)
	if err != nil {
		return buildErrorResponse(err), err
	}
	return buildResponse(districts, len(districts), total, page, limit), nil
}

func (u *geoUsecase) GetVillagesByDistrict(ctx context.Context, districtID, page, limit int) (domain.StandardResponse, error) {
	offset := calculateOffset(page, limit)
	villages, total, err := u.geoRepo.FetchVillagesByDistrict(ctx, districtID, limit, offset)
	if err != nil {
		return buildErrorResponse(err), err
	}
	return buildResponse(villages, len(villages), total, page, limit), nil
}
