package models

import "gorm.io/gorm"

type TenantRepository interface {
	Create(tenant *Tenant) error
	GetByID(id int) (*Tenant, error)
	List() ([]Tenant, error)
	Update(id int, tenant *Tenant) error
	Delete(id int) error
}

type TenantImpl struct {
	DB *gorm.DB
}

func NewTenantImpl(db *gorm.DB) TenantRepository {
	return &TenantImpl{DB: db}
}

func (t *TenantImpl) Create(tenant *Tenant) error {
	return t.DB.Create(tenant).Error
}

func (t *TenantImpl) GetByID(id int) (*Tenant, error) {
	var tenant Tenant
	if err := t.DB.First(&tenant, id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (t *TenantImpl) List() ([]Tenant, error) {
	var tenants []Tenant
	if err := t.DB.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (t *TenantImpl) Update(id int, tenant *Tenant) error {
	return t.DB.Model(&Tenant{}).Where("id = ?", id).Updates(tenant).Error
}

func (t *TenantImpl) Delete(id int) error {
	return t.DB.Delete(&Tenant{}, id).Error
} 