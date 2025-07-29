package models

import "gorm.io/gorm"

type LandLordRepository interface {
	Create(landLord *LandLord) error
	GetByID(id int) (*LandLord, error)
	List() ([]LandLord, error)
	Update(id int, landLord *LandLord) error
	Delete(id int) error
}

type LandLordImpl struct {
	DB *gorm.DB
}

func NewLandLordImpl(db *gorm.DB) LandLordRepository {
	return &LandLordImpl{DB: db}
}

func (l *LandLordImpl) Create(landLord *LandLord) error {
	return l.DB.Create(landLord).Error
}

func (l *LandLordImpl) GetByID(id int) (*LandLord, error) {
	var landLord LandLord
	if err := l.DB.First(&landLord, id).Error; err != nil {
		return nil, err
	}
	return &landLord, nil
}

func (l *LandLordImpl) List() ([]LandLord, error) {
	var landLords []LandLord
	if err := l.DB.Find(&landLords).Error; err != nil {
		return nil, err
	}
	return landLords, nil
}

func (l *LandLordImpl) Update(id int, landLord *LandLord) error {
	return l.DB.Model(&LandLord{}).Where("id = ?", id).Updates(landLord).Error
}

func (l *LandLordImpl) Delete(id int) error {
	return l.DB.Delete(&LandLord{}, id).Error
} 