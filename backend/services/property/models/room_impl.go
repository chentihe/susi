package models

import "gorm.io/gorm"

type RoomRepository interface {
	Create(room *Room) error
	GetByID(id int) (*Room, error)
	List() ([]Room, error)
	Update(id int, room *Room) error
	Delete(id int) error
}

type RoomImpl struct {
	DB *gorm.DB
}

func NewRoomImpl(db *gorm.DB) RoomRepository {
	return &RoomImpl{DB: db}
}

func (r *RoomImpl) Create(room *Room) error {
	return r.DB.Create(room).Error
}

func (r *RoomImpl) GetByID(id int) (*Room, error) {
	var room Room
	if err := r.DB.First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomImpl) List() ([]Room, error) {
	var rooms []Room
	if err := r.DB.Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *RoomImpl) Update(id int, room *Room) error {
	return r.DB.Model(&Room{}).Where("id = ?", id).Updates(room).Error
}

func (r *RoomImpl) Delete(id int) error {
	return r.DB.Delete(&Room{}, id).Error
} 