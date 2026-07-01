// Package model defines GORM models for the Order domain.
package model

import "time"

// MainOrder represents a customer order.
type MainOrder struct {
	ID                    int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo               string     `gorm:"column:order_no;size:30;uniqueIndex;not null" json:"order_no"`
	UserID                int64      `gorm:"column:user_id;not null;index:idx_order_user_status" json:"user_id"`
	ProductID             int64      `gorm:"column:product_id;not null;index" json:"product_id"`
	DepartureID           int64      `gorm:"column:departure_id;not null;index" json:"departure_id"`
	OrderStatus           string     `gorm:"column:order_status;size:30;not null;default:pending_pay;index:idx_order_user_status" json:"order_status"`
	PaymentStatus         string     `gorm:"column:payment_status;size:30;not null;default:unpaid" json:"payment_status"`
	TotalAmount           int64      `gorm:"column:total_amount;not null" json:"total_amount"`                         // cents
	DiscountAmount        int64      `gorm:"column:discount_amount;not null;default:0" json:"discount_amount"`         // cents
	PayableAmount         int64      `gorm:"column:payable_amount;not null" json:"payable_amount"`                     // cents
	AdultCount            int        `gorm:"column:adult_count;not null" json:"adult_count"`
	ChildCount            int        `gorm:"column:child_count;not null;default:0" json:"child_count"`
	InfantCount           int        `gorm:"column:infant_count;not null;default:0" json:"infant_count"`
	SingleSupplementAmount int64     `gorm:"column:single_supplement_amount;not null;default:0" json:"single_supplement_amount"` // cents
	AddonAmount           int64      `gorm:"column:addon_amount;not null;default:0" json:"addon_amount"`              // cents
	ContactName           string     `gorm:"column:contact_name;size:100;not null" json:"contact_name"`
	ContactPhone          string     `gorm:"column:contact_phone;size:20;not null" json:"contact_phone"`
	Channel               string     `gorm:"column:channel;size:20;not null;default:web" json:"channel"`
	Remark                string     `gorm:"column:remark;size:500" json:"remark,omitempty"`
	PaidAt                *time.Time `gorm:"column:paid_at" json:"paid_at,omitempty"`
	DepartedAt            *time.Time `gorm:"column:departed_at" json:"departed_at,omitempty"`
	CompletedAt           *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"`
	CancelledAt           *time.Time `gorm:"column:cancelled_at" json:"cancelled_at,omitempty"`
	CancelReason          string     `gorm:"column:cancel_reason;size:500" json:"cancel_reason,omitempty"`
	CreatedAt             time.Time  `gorm:"column:created_at;not null;default:now();index:idx_order_user_status" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`

	// Payment mode extension (FR-163, migration 007)
	PaymentMode    string     `gorm:"column:payment_mode;size:20;default:full" json:"payment_mode"`               // full/deposit
	DepositAmount  int64      `gorm:"column:deposit_amount" json:"deposit_amount"`                                 // cents
	BalanceAmount  int64      `gorm:"column:balance_amount" json:"balance_amount"`                                 // cents
	BalanceDeadline *time.Time `gorm:"column:balance_deadline" json:"balance_deadline,omitempty"`                  // 尾款截止时间
	DepositPaidAt  *time.Time `gorm:"column:deposit_paid_at" json:"deposit_paid_at,omitempty"`                     // 定金支付时间
	BalancePaidAt  *time.Time `gorm:"column:balance_paid_at" json:"balance_paid_at,omitempty"`                     // 尾款支付时间

	// Distribution tracking (migration 007)
	DistributorIDL1  *int64  `gorm:"column:distributor_id_l1" json:"distributor_id_l1,omitempty"`
	DistributorIDL2  *int64  `gorm:"column:distributor_id_l2" json:"distributor_id_l2,omitempty"`
	PromotionCode    string  `gorm:"column:promotion_code;size:20" json:"promotion_code,omitempty"`

	// Marketing (migration 007)
	CouponClaimID    *int64  `gorm:"column:coupon_claim_id" json:"coupon_claim_id,omitempty"`
	CouponDiscount   int64   `gorm:"column:coupon_discount" json:"coupon_discount"`                               // cents
	ActivityID       *int64  `gorm:"column:activity_id" json:"activity_id,omitempty"`
	ActivityDiscount int64   `gorm:"column:activity_discount" json:"activity_discount"`                           // cents

	// Relations
	SubOrders      []SubOrder      `gorm:"foreignKey:MainOrderID" json:"sub_orders,omitempty"`
	StatusLogs     []OrderStatusLog `gorm:"foreignKey:OrderID" json:"status_logs,omitempty"`
	Travellers     []OrderTraveller `gorm:"foreignKey:OrderID" json:"travellers,omitempty"`
}

// TableName overrides the table name.
func (MainOrder) TableName() string {
	return "main_order"
}

// SubOrder represents a sub-order for ancillary services (insurance, transfer).
type SubOrder struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MainOrderID   int64     `gorm:"column:main_order_id;not null;index" json:"main_order_id"`
	SubOrderNo    string    `gorm:"column:sub_order_no;size:30;uniqueIndex;not null" json:"sub_order_no"`
	ResourceType  string    `gorm:"column:resource_type;size:30;not null" json:"resource_type"`
	ResourceID    *int64    `gorm:"column:resource_id" json:"resource_id,omitempty"`
	ResourceName  string    `gorm:"column:resource_name;size:200;not null" json:"resource_name"`
	SupplierID    *int64    `gorm:"column:supplier_id" json:"supplier_id,omitempty"`
	Status        string    `gorm:"column:status;size:20;not null;default:pending" json:"status"`
	Amount        int64     `gorm:"column:amount;not null" json:"amount"` // cents
	CommissionRate float64  `gorm:"column:commission_rate;type:decimal(5,4);default:0" json:"commission_rate"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (SubOrder) TableName() string {
	return "sub_order"
}

// OrderStatusLog records order status transitions for audit trail.
type OrderStatusLog struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID      int64     `gorm:"column:order_id;not null;index" json:"order_id"`
	FromStatus   string    `gorm:"column:from_status;size:30;not null" json:"from_status"`
	ToStatus     string    `gorm:"column:to_status;size:30;not null" json:"to_status"`
	OperatorType string    `gorm:"column:operator_type;size:20;not null" json:"operator_type"`
	OperatorID   *int64    `gorm:"column:operator_id" json:"operator_id,omitempty"`
	Reason       string    `gorm:"column:reason;size:500" json:"reason,omitempty"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (OrderStatusLog) TableName() string {
	return "order_status_log"
}

// OrderTraveller represents a traveller associated with an order.
type OrderTraveller struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID       int64      `gorm:"column:order_id;not null;index" json:"order_id"`
	RealName      string     `gorm:"column:real_name;type:text;not null" json:"-"` // AES-256-GCM encrypted
	IDCardNo      string     `gorm:"column:id_card_no;type:text;not null" json:"-"` // AES-256-GCM encrypted
	Phone         string     `gorm:"column:phone;size:20" json:"phone"`
	BirthDate     *time.Time `gorm:"column:birth_date" json:"birth_date,omitempty"`
	Gender        string     `gorm:"column:gender;size:10" json:"gender"`
	IsChild       bool       `gorm:"column:is_child;not null;default:false" json:"is_child"`
	IsInfant      bool       `gorm:"column:is_infant;not null;default:false" json:"is_infant"`
	LinkedAdultID *int64     `gorm:"column:linked_adult_id" json:"linked_adult_id,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name.
func (OrderTraveller) TableName() string {
	return "order_traveller"
}

// Order status constants (internal states, snake_case).
const (
	OrderStatusPendingPay    = "pending_pay"
	OrderStatusPaidDeposit   = "paid_deposit"    // FR-164: 定金已付
	OrderStatusPendingBalance = "pending_balance" // FR-164: 待付尾款
	OrderStatusPaidFull      = "paid_full"
	OrderStatusPendingTravel = "pending_travel"
	OrderStatusInTravel      = "in_travel"
	OrderStatusCompleted     = "completed"
	OrderStatusCancelled     = "cancelled"
	OrderStatusRefunding     = "refunding"
	OrderStatusRefunded      = "refunded"
	OrderStatusClosed        = "closed"
)

// Payment status constants.
const (
	PaymentStatusUnpaid    = "unpaid"
	PaymentStatusPartial   = "partial"
	PaymentStatusPaid      = "paid"
	PaymentStatusRefunded  = "refunded"
)

// ValidTransitions defines the allowed order status transitions.
var ValidTransitions = map[string][]string{
	OrderStatusPendingPay:     {OrderStatusPaidFull, OrderStatusPaidDeposit, OrderStatusCancelled, OrderStatusRefunding},
	OrderStatusPaidDeposit:    {OrderStatusPendingBalance, OrderStatusCancelled, OrderStatusRefunding},
	OrderStatusPendingBalance: {OrderStatusPaidFull, OrderStatusCancelled},
	OrderStatusPaidFull:       {OrderStatusPendingTravel, OrderStatusRefunding},
	OrderStatusPendingTravel:  {OrderStatusInTravel},
	OrderStatusInTravel:       {OrderStatusCompleted},
	OrderStatusCompleted:      {OrderStatusClosed},
	OrderStatusCancelled:      {OrderStatusClosed},
	OrderStatusRefunding:      {OrderStatusRefunded, OrderStatusPaidFull},
	OrderStatusRefunded:       {OrderStatusClosed},
}

// CanTransitionTo checks if the order can transition from its current status to the target.
func CanTransitionTo(current, target string) bool {
	allowed, ok := ValidTransitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

// Channel constants.
const (
	ChannelWeb    = "web"
	ChannelMiniApp = "miniapp"
	ChannelAdmin  = "admin"
)

// Payment mode constants.
const (
	PaymentModeFull    = "full"    // 全额支付
	PaymentModeDeposit = "deposit" // 定金+尾款
)

// Default deposit configuration.
const (
	DefaultDepositRatio = 0.30 // 默认定金比例 30%
	MinDepositRatio     = 0.10 // 最小定金比例 10%
	MaxDepositRatio     = 0.50 // 最大定金比例 50%
	DefaultGracePeriodHours = 24 // 默认宽限期 24 小时
	DefaultReminderDaysBefore = 3 // 默认提前提醒天数
)
