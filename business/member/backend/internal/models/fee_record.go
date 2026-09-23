package models

// FeeStatus constants
const (
	FeeStatusUnpaid  = "unpaid"  // 未缴费
	FeeStatusPending = "pending" // 待确认
	FeeStatusPaid    = "paid"    // 已缴费
)

// FeeRecord represents annual membership fee record.
//
// 不变量：同一会员同一年度至多一条记录。该约束由 `uk_member_year(member_id, year)` 唯一索引保证，
// 启动时由 `db.EnsureUniqueMemberFeeIndex()` 幂等创建（**故意**不在此处声明 uniqueIndex：
// AutoMigrate 在既有表上补建唯一索引遇到历史重复数据会直接报错退出）。
type FeeRecord struct {
	BaseModel
	MemberID      uint64     `gorm:"index;not null" json:"member_id"`
	Member        Member     `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Year          int        `gorm:"not null" json:"year"`
	Amount        float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status        string     `gorm:"size:20;default:unpaid" json:"status"`
	PaidAt        *LocalTime `gorm:"type:datetime" json:"paid_at"`
	TransactionID string     `gorm:"size:128" json:"transaction_id"`
	InvoiceNo     string     `gorm:"size:64" json:"invoice_no"`
	Remark        string     `gorm:"size:500" json:"remark"`
	FeeStandardID uint64     `gorm:"default:0" json:"fee_standard_id"`
	LevelID       uint64     `gorm:"default:0" json:"level_id"`
	LevelName     string     `gorm:"size:64" json:"level_name"`
	OrgID         uint64     `gorm:"default:0" json:"org_id"`
	OrgName       string     `gorm:"size:255" json:"org_name"`
	ReceiptFile   string     `gorm:"size:255" json:"receipt_file"`                    // 缴费回执单 URL
	PaidDate      string     `gorm:"size:10" json:"paid_date"`                        // 缴费日期 YYYY-MM-DD
	PaidAmount    float64    `gorm:"type:decimal(10,2);default:0" json:"paid_amount"` // 实际缴费金额
	ConfirmedAt   *LocalTime `gorm:"type:datetime" json:"confirmed_at"`               // 管理员确认时间

	// Invoice fields
	InvoiceStatus   string     `gorm:"size:20;default:''" json:"invoice_status"`           // 开票状态: ""=未申请, "applied"=已申请, "issued"=已开票
	InvoiceCompany  string     `gorm:"size:255;default:''" json:"invoice_company"`         // 开票单位全称
	InvoiceTaxID    string     `gorm:"size:64;default:''" json:"invoice_tax_id"`           // 开票单位统一社会信用代码
	InvoiceAmount   float64    `gorm:"type:decimal(10,2);default:0" json:"invoice_amount"` // 开票金额
	InvoiceContact  string     `gorm:"size:128;default:''" json:"invoice_contact"`         // 开票联系人及电话
	InvoiceRemark   string     `gorm:"size:500;default:''" json:"invoice_remark"`          // 开票备注
	InvoiceFile     string     `gorm:"size:255;default:''" json:"invoice_file"`            // 发票 PDF 文件 URL
	InvoiceIssuedAt *LocalTime `gorm:"type:datetime" json:"invoice_issued_at"`             // 开票时间
}

func (FeeRecord) TableName() string {
	return "member_fee_records"
}
