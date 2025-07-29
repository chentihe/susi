package models

import "gorm.io/gorm"

type AdminRepository interface {
	Create(admin *Admin) error
	GetByID(id int) (*Admin, error)
	List() ([]Admin, error)
	Update(id int, admin *Admin) error
	Delete(id int) error
}

type AdminImpl struct {
	DB *gorm.DB
}

func NewAdminImpl(db *gorm.DB) AdminRepository {
	return &AdminImpl{DB: db}
}

func (a *AdminImpl) Create(admin *Admin) error {
	return a.DB.Create(admin).Error
}

func (a *AdminImpl) GetByID(id int) (*Admin, error) {
	var admin Admin
	if err := a.DB.First(&admin, id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (a *AdminImpl) List() ([]Admin, error) {
	var admins []Admin
	if err := a.DB.Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (a *AdminImpl) Update(id int, admin *Admin) error {
	return a.DB.Model(&Admin{}).Where("id = ?", id).Updates(admin).Error
}

func (a *AdminImpl) Delete(id int) error {
	return a.DB.Delete(&Admin{}, id).Error
} 