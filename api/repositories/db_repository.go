package repositories

import (
	"context"
	"math"

	"gorm.io/gorm"
)

type IDBRepository[T any] interface {
	Create(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id uint) (*T, error)
	GetAll(ctx context.Context) ([]T, error)
	GetPaginated(ctx context.Context, page, limit int, search string) (*PaginatedResponse[T], error)
	FindByIDWithAssociations(ctx context.Context, id uint, associations ...string) (*T, error)
	FindWithAssociations(ctx context.Context, query string, associations []string, args ...any) ([]T, error)
	Find(ctx context.Context, query string, args ...any) ([]T, error)
	Count(ctx context.Context) (int64, error)
	Save(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uint) error
}

type dbRepository[T any] struct {
	db *gorm.DB
}

func NewDBRepository[T any](db *gorm.DB) IDBRepository[T] {
	return &dbRepository[T]{
		db: db,
	}
}

func (r *dbRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *dbRepository[T]) GetAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func (r *dbRepository[T]) GetPaginated(ctx context.Context, page, limit int, search string) (*PaginatedResponse[T], error) {
	var items []T
	var totalItems int64

	dbQuery := r.db.WithContext(ctx).Model(new(T))

	if search != "" {
		searchTerm := "%" + search + "%"
		dbQuery = dbQuery.Where(
			"title ILIKE ? OR authors ILIKE ? OR genre ILIKE ? OR publisher ILIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	if err := dbQuery.Count(&totalItems).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * limit

	if err := dbQuery.Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	return &PaginatedResponse[T]{
		Data:       items,
		Page:       page,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}

func (r *dbRepository[T]) FindByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *dbRepository[T]) FindByIDWithAssociations(ctx context.Context, id uint, associations ...string) (*T, error) {
	var entity T

	query := r.db.WithContext(ctx)

	for _, assoc := range associations {
		query = query.Preload(assoc)
	}

	err := query.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *dbRepository[T]) Find(ctx context.Context, query string, args ...any) ([]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Where(query, args...).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *dbRepository[T]) FindWithAssociations(ctx context.Context, query string, associations []string, args ...any) ([]T, error) {
	var entities []T

	dbQuery := r.db.WithContext(ctx)

	for _, assoc := range associations {
		dbQuery = dbQuery.Preload(assoc)
	}

	dbQuery = dbQuery.Where(query, args...)

	err := dbQuery.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *dbRepository[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *dbRepository[T]) Save(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *dbRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}
