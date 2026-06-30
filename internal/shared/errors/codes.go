// Package errors defines unified error codes for all microservices.
// All API responses use these codes in the {code, message, data, trace_id} envelope.
//
// Code ranges:
//   0     : Success
//   1000-1999: Client errors (bad request, unauthorized, not found, etc.)
//   2000-2999: Business logic errors (insufficient balance, order closed, etc.)
//   3000-3999: Domain-specific errors (supplier, distribution, visa, etc.)
//   5000-5999: Server errors (database, cache, third-party, etc.)
package errors

import "fmt"

// AppError represents a structured application error with a code and message.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New creates a new AppError.
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// ────────────────────────────────────────────────────────────────────────────
// General Client Errors (1000-1999)
// ────────────────────────────────────────────────────────────────────────────

const (
	CodeSuccess      = 0
	CodeBadRequest   = 1001
	CodeUnauthorized = 1002
	CodeForbidden    = 1003
	CodeNotFound     = 1004
	CodeValidation   = 1005
	CodeConflict     = 1006
	CodeTooManyReq   = 1007
	CodeTimeout      = 1008
)

// ────────────────────────────────────────────────────────────────────────────
// Business Logic Errors (2000-2999)
// ────────────────────────────────────────────────────────────────────────────

const (
	CodeBusiness      = 2000
	CodeInsufficient  = 2001
	CodePaymentFailed = 2002
	CodeOrderClosed   = 2003
	CodeStockEmpty    = 2004
	CodeOrderTimeout  = 2005
	CodeRefundExceed  = 2006
)

// ────────────────────────────────────────────────────────────────────────────
// Domain-Specific Errors (3000-3999)
// ────────────────────────────────────────────────────────────────────────────

// Supplier domain errors (3000-3099)
const (
	CodeSupplierPending      = 3000
	CodeSupplierSuspended    = 3001
	CodeSupplierTerminated   = 3002
	CodeDuplicateCreditCode  = 3003
	CodeSettlementDisputed   = 3004
	CodeWithdrawalMinAmount  = 3005
)

// Distribution domain errors (3100-3199)
const (
	CodeDistributorPending   = 3100
	CodeDistributorFrozen    = 3101
	CodeDistributorCancelled = 3102
	CodeCommissionFrozen     = 3103
	CodeWithdrawalPending    = 3104
	CodeAntiFraudBlocked     = 3105
	CodeSelfPurchaseBan      = 3106
	CodeDuplicateInviteCode  = 3107
)

// Visa domain errors (3200-3299)
const (
	CodeVisaOrderNotFound    = 3200
	CodeVisaInvalidTransition = 3201
	CodeVisaMaterialTooLarge = 3202
	CodeVisaMaterialRejected = 3203
	CodePassportExpiring     = 3204
)

// Marketing domain errors (3300-3399)
const (
	CodeCouponExhausted    = 3300
	CodeCouponExpired      = 3301
	CodeCouponLimitReached = 3302
	CodeCouponNotApplicable = 3303
	CodeActivityNotActive  = 3304
	CodeActivityStockEmpty = 3305
)

// Payment extension errors (3400-3499)
const (
	CodeDepositRequired    = 3400
	CodeBalanceOverdue     = 3401
	CodePartialRefundExceed = 3402
	CodeUnionPayError      = 3403
)

// ────────────────────────────────────────────────────────────────────────────
// Server Errors (5000-5999)
// ────────────────────────────────────────────────────────────────────────────

const (
	CodeServer    = 5000
	CodeDBError   = 5001
	CodeCacheError = 5002
	CodeThirdParty = 5003
	CodeNATSError = 5004
	CodeConsulError = 5005
	CodeMeiliError  = 5006
)

// ────────────────────────────────────────────────────────────────────────────
// Convenience constructors
// ────────────────────────────────────────────────────────────────────────────

func BadRequest(msg string) *AppError     { return New(CodeBadRequest, msg) }
func Unauthorized(msg string) *AppError   { return New(CodeUnauthorized, msg) }
func Forbidden(msg string) *AppError      { return New(CodeForbidden, msg) }
func NotFound(msg string) *AppError       { return New(CodeNotFound, msg) }
func Conflict(msg string) *AppError       { return New(CodeConflict, msg) }
func TooManyRequests(msg string) *AppError { return New(CodeTooManyReq, msg) }
func ServerErr(msg string) *AppError      { return New(CodeServer, msg) }
func DBErr(msg string) *AppError          { return New(CodeDBError, msg) }
