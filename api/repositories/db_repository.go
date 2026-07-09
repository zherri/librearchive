package repositories

import (
	"context"

	"gorm.io/gorm"
)

type IDBRepository[T any] interface {
	Create(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id uint) (*T, error)
	FindAll(ctx context.Context) ([]T, error)
	FindByIDWithAssociations(ctx context.Context, id uint, associations ...string) (*T, error)
	FindWithAssociations(ctx context.Context, queryCond string, associations []string, args ...any) ([]T, error)
	Find(ctx context.Context, query string, args ...any) ([]T, error)
	Save(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uint) error
}

type db_repository[T any] struct {
	db *gorm.DB
}

func NewDBRepository[T any](db *gorm.DB) IDBRepository[T] {
	return &db_repository[T]{
		db: db,
	}
}

func (dbr *db_repository[T]) Create(ctx context.Context, entity *T) error {
	return dbr.db.WithContext(ctx).Create(entity).Error
}

func (dbr *db_repository[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := dbr.db.WithContext(ctx).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (dbr *db_repository[T]) FindByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	err := dbr.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (dbr *db_repository[T]) FindByIDWithAssociations(ctx context.Context, id uint, associations ...string) (*T, error) {
	var entity T

	query := dbr.db.WithContext(ctx)

	for _, assoc := range associations {
		query = query.Preload(assoc)
	}

	err := query.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (dbr *db_repository[T]) Find(ctx context.Context, query string, args ...any) ([]T, error) {
	var entities []T
	err := dbr.db.WithContext(ctx).Where(query, args...).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (dbr *db_repository[T]) FindWithAssociations(ctx context.Context, query string, associations []string, args ...any) ([]T, error) {
	var entities []T

	dbQuery := dbr.db.WithContext(ctx)

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

func (dbr *db_repository[T]) Save(ctx context.Context, entity *T) error {
	return dbr.db.WithContext(ctx).Save(entity).Error
}

func (dbr *db_repository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	return dbr.db.WithContext(ctx).Delete(&entity, id).Error
}
