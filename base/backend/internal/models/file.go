package models

// UploadedFile 上传文件元数据
type UploadedFile struct {
	BaseModel
	TenantID uint64 `gorm:"index;comment:租户ID" json:"tenantId"`
	UserID   uint64 `gorm:"index;comment:用户ID" json:"userId"`
	FileName string `gorm:"size:256;comment:原始文件名" json:"fileName"`
	FileKey  string `gorm:"size:256;uniqueIndex;comment:存储key" json:"fileKey"`
	FileType string `gorm:"size:64;comment:文件类型" json:"fileType"`
	FileSize int64  `gorm:"comment:文件大小(字节)" json:"fileSize"`
	URL      string `gorm:"size:512;comment:访问地址" json:"url"`
	Storage  string `gorm:"size:32;comment:存储方式 local" json:"storage"`
}

func (UploadedFile) TableName() string {
	return "base_uploaded_file"
}
