package ds

type User struct {
	ID          uint   `gorm:"primaryKey"`
	Login       string `gorm:"type:varchar(50);unique;not null"`
	Password    string `gorm:"type:varchar(100);not null"`
	IsModerator bool   `gorm:"type:boolean;default:false"`
	Name        string `gorm:"type:varchar(100)"`
	Email       string `gorm:"type:varchar(100)"`
	Phone       string `gorm:"type:varchar(20)"`
}