package models

import (
	"gorm.io/gorm"
)

type ApartmentRepository interface {
	Create(apartment *Apartment) error
	GetByID(id int) (*Apartment, error)
	List() ([]Apartment, error)
	Update(id int, apartment *Apartment) error
	Delete(id int) error
}

type ApartmentImpl struct {
	DB *gorm.DB
}

func NewApartmentImpl(db *gorm.DB) ApartmentRepository {
	return &ApartmentImpl{DB: db}
}

func (a *ApartmentImpl) Create(apartment *Apartment) error {
	return a.DB.Create(apartment).Error
}

func (a *ApartmentImpl) GetByID(id int) (*Apartment, error) {
	var apt Apartment
	if err := a.DB.First(&apt, id).Error; err != nil {
		return nil, err
	}
	return &apt, nil
}

func (a *ApartmentImpl) List() ([]Apartment, error) {
	var apartments []Apartment
	if err := a.DB.Find(&apartments).Error; err != nil {
		return nil, err
	}
	return apartments, nil
}

func (a *ApartmentImpl) Update(id int, apartment *Apartment) error {
	return a.DB.Model(&Apartment{}).Where("id = ?", id).Updates(apartment).Error
}

func (a *ApartmentImpl) Delete(id int) error {
	return a.DB.Delete(&Apartment{}, id).Error
} 