package models

import "gorm.io/gorm"

type PropertyRepository interface {
	Create(property *Property) error
	GetByID(id int) (*Property, error)
	List() ([]Property, error)
	Update(id int, property *Property) error
	Delete(id int) error
}

type PropertyImpl struct {
	DB *gorm.DB
}

func NewPropertyImpl(db *gorm.DB) PropertyRepository {
	return &PropertyImpl{DB: db}
}

func (p *PropertyImpl) Create(property *Property) error {
	return p.DB.Create(property).Error
}

func (p *PropertyImpl) GetByID(id int) (*Property, error) {
	var prop Property
	if err := p.DB.First(&prop, id).Error; err != nil {
		return nil, err
	}
	return &prop, nil
}

func (p *PropertyImpl) List() ([]Property, error) {
	var properties []Property
	if err := p.DB.Find(&properties).Error; err != nil {
		return nil, err
	}
	return properties, nil
}

func (p *PropertyImpl) Update(id int, property *Property) error {
	return p.DB.Model(&Property{}).Where("id = ?", id).Updates(property).Error
}

func (p *PropertyImpl) Delete(id int) error {
	return p.DB.Delete(&Property{}, id).Error
} 