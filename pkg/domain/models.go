package domain

type Region struct {
	ID          int64   `gorm:"primaryKey" json:"id"`
	KorName     string  `json:"kor_name"`
	EngName     string  `json:"eng_name"`
	Continent   string  `json:"continent"`
	Description *string `json:"description"`
}

type Distillery struct {
	ID          int64   `gorm:"primaryKey" json:"id"`
	KorName     string  `json:"kor_name"`
	EngName     string  `json:"eng_name"`
	LogoImgUrl  *string `json:"logo_img_url"`
	Description *string `json:"description"`
}

type TastingTag struct {
	ID          int64   `gorm:"primaryKey" json:"id"`
	KorName     string  `json:"kor_name"`
	EngName     string  `json:"eng_name"`
	Icon        *string `json:"icon"`
	Description *string `json:"description"`
}

type Alcohol struct {
	ID            int64        `gorm:"primaryKey" json:"id"`
	KorName       string       `json:"kor_name"`
	EngName       string       `json:"eng_name"`
	Type          string       `json:"type"`
	ABV           *string      `json:"abv"`
	Volume        *string      `json:"volume"`
	Age           *string      `json:"age"`
	Cask          *string      `json:"cask"`
	KorCategory   *string      `json:"kor_category"`
	EngCategory   *string      `json:"eng_category"`
	CategoryGroup *string      `json:"category_group"`
	ImageURL      *string      `json:"image_url"`
	Description   *string      `json:"description"`
	RegionID      *int64       `json:"region_id"`
	DistilleryID  *int64       `json:"distillery_id"`
	Region        *Region      `gorm:"foreignKey:RegionID" json:"region"`
	Distillery    *Distillery  `gorm:"foreignKey:DistilleryID" json:"distillery"`
	TastingTags   []TastingTag `gorm:"many2many:alcohol_tasting_tags" json:"tasting_tags"`
}
