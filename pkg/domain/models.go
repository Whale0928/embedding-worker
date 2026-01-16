package domain

type Region struct {
	ID          int64   `gorm:"primaryKey" json:"id"`
	KorName     string  `json:"kor_name"`
	EngName     string  `json:"eng_name"`
	Continent   string  `json:"continent"`
	Description *string `json:"description"`
}
