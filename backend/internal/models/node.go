package models

type Node struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"uniqueIndex"`
	Address string
	Port    int
	Active  bool `gorm:"default:true"`
}
