package models

import "gorm.io/gorm"

type RenovationRepository interface {
	Create(renovation *Renovation) error
	GetByID(id int) (*Renovation, error)
	List() ([]Renovation, error)
	Update(id int, renovation *Renovation) error
	Delete(id int) error
}

type RenovationImpl struct {
	DB *gorm.DB
}

func NewRenovationImpl(db *gorm.DB) RenovationRepository {
	return &RenovationImpl{DB: db}
}

func (r *RenovationImpl) Create(renovation *Renovation) error {
	return r.DB.Create(renovation).Error
}

func (r *RenovationImpl) GetByID(id int) (*Renovation, error) {
	var renovation Renovation
	if err := r.DB.First(&renovation, id).Error; err != nil {
		return nil, err
	}
	return &renovation, nil
}

func (r *RenovationImpl) List() ([]Renovation, error) {
	var renovations []Renovation
	if err := r.DB.Find(&renovations).Error; err != nil {
		return nil, err
	}
	return renovations, nil
}

func (r *RenovationImpl) Update(id int, renovation *Renovation) error {
	return r.DB.Model(&Renovation{}).Where("id = ?", id).Updates(renovation).Error
}

func (r *RenovationImpl) Delete(id int) error {
	return r.DB.Delete(&Renovation{}, id).Error
} 