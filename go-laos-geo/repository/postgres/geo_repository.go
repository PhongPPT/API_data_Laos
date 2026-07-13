package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/user/go-laos-geo/domain"
)

type geoRepository struct {
	db *sqlx.DB
}

func NewGeoRepository(db *sqlx.DB) domain.GeoRepository {
	return &geoRepository{db}
}

func (r *geoRepository) FetchProvinces(ctx context.Context, limit, offset int) ([]domain.Province, int, error) {
	var provinces []domain.Province
	var total int

	countQuery := `SELECT COUNT(*) FROM provinces`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT pr_id, pr_name, pr_name_en, lat, lng FROM provinces LIMIT $1 OFFSET $2`
	err = r.db.SelectContext(ctx, &provinces, query, limit, offset)
	return provinces, total, err
}

func (r *geoRepository) FetchDistrictsByProvince(ctx context.Context, provinceID, limit, offset int) ([]domain.District, int, error) {
	var districts []domain.District
	var total int

	countQuery := `SELECT COUNT(*) FROM districts WHERE pr_id = $1`
	err := r.db.GetContext(ctx, &total, countQuery, provinceID)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT dr_id, dr_name, dr_name_en, pr_id, lat, lng FROM districts WHERE pr_id = $1 LIMIT $2 OFFSET $3`
	err = r.db.SelectContext(ctx, &districts, query, provinceID, limit, offset)
	return districts, total, err
}

func (r *geoRepository) FetchVillagesByDistrict(ctx context.Context, districtID, limit, offset int) ([]domain.Village, int, error) {
	var villages []domain.Village
	var total int

	countQuery := `SELECT COUNT(*) FROM villages WHERE dr_id = $1`
	err := r.db.GetContext(ctx, &total, countQuery, districtID)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT vill_id, vill_name, vill_name_en, dr_id, lat, lng FROM villages WHERE dr_id = $1 LIMIT $2 OFFSET $3`
	err = r.db.SelectContext(ctx, &villages, query, districtID, limit, offset)
	return villages, total, err
}
