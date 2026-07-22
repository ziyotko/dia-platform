package models

import "gorm.io/gorm"

// Dict 数据字典分组
type Dict struct {
	BaseModel
	Code        string `gorm:"size:64;uniqueIndex;comment:字典编码" json:"code"`
	Name        string `gorm:"size:128;comment:字典名称" json:"name"`
	Description string `gorm:"size:512;comment:描述" json:"description"`
	Status      int    `gorm:"default:1;comment:状态" json:"status"`
	Items       []DictItem `json:"items" gorm:"foreignKey:DictID"`
}

func (Dict) TableName() string {
	return "base_dict"
}

// DictItem 数据字典项
type DictItem struct {
	BaseModel
	DictID uint64 `gorm:"index;comment:字典ID" json:"dictId"`
	Label  string `gorm:"size:128;comment:显示名" json:"label"`
	Value  string `gorm:"size:128;comment:字典值" json:"value"`
	Sort   int    `gorm:"default:0;comment:排序" json:"sort"`
	Status int    `gorm:"default:1;comment:状态" json:"status"`
}

func (DictItem) TableName() string {
	return "base_dict_item"
}

func AutoMigrateDict(db *gorm.DB) error {
	return db.AutoMigrate(&Dict{}, &DictItem{})
}
