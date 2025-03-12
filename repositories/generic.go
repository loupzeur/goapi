package repositories

import (
	"context"

	"gorm.io/gorm"
)

type Repository[T any] interface {
	Create(context.Context, *T, ...func(db *gorm.DB) *gorm.DB) error
	Update(context.Context, *T, ...func(*gorm.DB) *gorm.DB) error
	Delete(context.Context, *T, ...func(*gorm.DB) *gorm.DB) error
	FindByID(context.Context, any, ...func(*gorm.DB) *gorm.DB) (*T, error)
	FindByScope(context.Context, ...func(*gorm.DB) *gorm.DB) (*T, error)
	FindAll(context.Context, ...func(*gorm.DB) *gorm.DB) ([]T, error)
}
type RepositoryGeneric[T any] struct {
	db *gorm.DB
}

func NewRepository[T any](db *gorm.DB) *RepositoryGeneric[T] {
	return &RepositoryGeneric[T]{db: db}
}
func (r *RepositoryGeneric[T]) Create(c context.Context, data *T, funcs ...func(db *gorm.DB) *gorm.DB) error {

	return r.db.WithContext(c).Scopes(funcs...).Save(data).Error
}
func (r *RepositoryGeneric[T]) Update(c context.Context, data *T, funcs ...func(db *gorm.DB) *gorm.DB) error {

	return r.db.WithContext(c).Scopes(funcs...).Updates(data).Error
}
func (r *RepositoryGeneric[T]) Delete(c context.Context, data *T, funcs ...func(db *gorm.DB) *gorm.DB) error {

	return r.db.WithContext(c).Scopes(funcs...).Delete(data).Error
}
func (r *RepositoryGeneric[T]) FindByID(c context.Context, id any, funcs ...func(db *gorm.DB) *gorm.DB) (*T, error) {

	data := new(T)
	err := r.db.WithContext(c).Scopes(funcs...).First(data, id).Error
	return data, err
}
func (r *RepositoryGeneric[T]) FindByScope(c context.Context, funcs ...func(db *gorm.DB) *gorm.DB) (*T, error) {

	data := new(T)
	err := r.db.WithContext(c).Scopes(funcs...).First(data).Error
	return data, err
}
func (r *RepositoryGeneric[T]) FindAll(c context.Context, funcs ...func(db *gorm.DB) *gorm.DB) ([]T, error) {

	data := new([]T)
	err := r.db.WithContext(c).Scopes(funcs...).Find(data).Error
	return *data, err
}
