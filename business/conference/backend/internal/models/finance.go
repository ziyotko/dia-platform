package models

import "time"

// Order for conference fees
type Order struct {
	BaseModel
	MeetingID     uint64     `gorm:"index" json:"meetingId"`
	UserID        uint64     `gorm:"index" json:"userId"`
	OrderNo       string     `gorm:"size:64;uniqueIndex" json:"orderNo"`
	Amount        float64    `gorm:"type:decimal(10,2)" json:"amount"`
	Status        string     `gorm:"size:32;default:unpaid" json:"status"` // unpaid/paid/refunding/refunded
	PayMethod     string     `gorm:"size:32" json:"payMethod"`             // wechat/alipay
	TransactionID string     `gorm:"size:128" json:"transactionId"`
	PaidAt        *time.Time `json:"paidAt"`

	// Relations
	Meeting *Meeting `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Order) TableName() string {
	return "conference_orders"
}

// Refund application
type Refund struct {
	BaseModel
	OrderID       uint64     `gorm:"index" json:"orderId"`
	UserID        uint64     `gorm:"index" json:"userId"`
	Amount        float64    `gorm:"type:decimal(10,2)" json:"amount"`
	Reason        string     `gorm:"size:512" json:"reason"`
	Status        string     `gorm:"size:32;default:pending" json:"status"` // pending/approved/rejected
	ReviewedBy    uint64     `gorm:"default:0" json:"reviewedBy"`
	ReviewedAt    *time.Time `json:"reviewedAt"`
	ReviewComment string     `gorm:"size:512" json:"reviewComment"`

	// Relations
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Refund) TableName() string {
	return "conference_refunds"
}

// Invoice information
type Invoice struct {
	BaseModel
	OrderID     uint64 `gorm:"index" json:"orderId"`
	UserID      uint64 `gorm:"index" json:"userId"`
	Type        string `gorm:"size:32" json:"type"` // individual/company
	Title       string `gorm:"size:256" json:"title"`
	TaxNo       string `gorm:"size:64" json:"taxNo"`
	Address     string `gorm:"size:512" json:"address"`
	Phone       string `gorm:"size:32" json:"phone"`
	Bank        string `gorm:"size:128" json:"bank"`
	BankAccount string `gorm:"size:64" json:"bankAccount"`

	// Relations
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

func (Invoice) TableName() string {
	return "conference_invoices"
}
