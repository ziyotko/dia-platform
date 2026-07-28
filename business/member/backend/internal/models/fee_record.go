package models

// FeeStatus constants
const (
	FeeStatusUnpaid = "unpaid" // 未缴费
	FeeStatusPaid   = "paid"   // 已缴费
)

// FeeRecord represents annual membership fee record
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
}

func (FeeRecord) TableName() string {
	return "member_fee_records"
}
