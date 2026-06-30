// Package service provides business logic for the Order domain.
package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/travel-booking/server/internal/order/model"
)

// OCRAdapter defines the interface for passport OCR recognition.
// Implementations should support passport photo scanning and field extraction.
type OCRAdapter interface {
	// RecognizePassport performs OCR on a passport image and returns extracted fields.
	RecognizePassport(ctx context.Context, imageReader io.Reader) (*model.OCRResult, error)
}

// BaiduOCRConfig holds configuration for Baidu OCR API.
type BaiduOCRConfig struct {
	APIKey    string
	SecretKey string
	Endpoint  string
}

// BaiduOCRAdapter implements OCRAdapter using Baidu's passport OCR API.
type BaiduOCRAdapter struct {
	config    BaiduOCRConfig
	accessToken string
	tokenExpiry time.Time
}

// NewBaiduOCRAdapter creates a new BaiduOCRAdapter.
func NewBaiduOCRAdapter(config BaiduOCRConfig) *BaiduOCRAdapter {
	if config.Endpoint == "" {
		config.Endpoint = "https://aip.baidubce.com"
	}
	return &BaiduOCRAdapter{
		config: config,
	}
}

// RecognizePassport performs OCR on a passport image using Baidu's API.
func (a *BaiduOCRAdapter) RecognizePassport(ctx context.Context, imageReader io.Reader) (*model.OCRResult, error) {
	// Read image data
	imageData, err := io.ReadAll(imageReader)
	if err != nil {
		return &model.OCRResult{
			Success:      false,
			ErrorMessage: fmt.Errorf("read image: %w", err).Error(),
		}, nil
	}

	// Ensure we have a valid access token
	if err := a.ensureToken(ctx); err != nil {
		return &model.OCRResult{
			Success:      false,
			ErrorMessage: fmt.Errorf("get access token: %w", err).Error(),
		}, nil
	}

	// Call Baidu passport OCR API
	result, err := a.callPassportOCR(ctx, imageData)
	if err != nil {
		return &model.OCRResult{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return result, nil
}

// ensureToken refreshes the access token if expired.
func (a *BaiduOCRAdapter) ensureToken(ctx context.Context) error {
	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return nil
	}

	// In production, this would call Baidu's token endpoint:
	// POST https://aip.baidubce.com/oauth/2.0/token
	// For now, this is a placeholder that should be configured via environment variables.
	return fmt.Errorf("baidu OCR access token not configured - set BAIDU_OCR_API_KEY and BAIDU_OCR_SECRET_KEY")
}

// callPassportOCR calls the Baidu passport OCR API.
func (a *BaiduOCRAdapter) callPassportOCR(ctx context.Context, imageData []byte) (*model.OCRResult, error) {
	// In production, this would call:
	// POST https://aip.baidubce.com/rest/2.0/ocr/v1/passport
	// with the image data as base64 encoded form data
	//
	// Response format:
	// {
	//   "words_result": {
	//     "Name": {"word": "张三"},
	//     "Country": {"word": "CHN"},
	//     "Sex": {"word": "M"},
	//     "BirthDate": {"word": "1990-01-01"},
	//     "ExpiryDate": {"word": "2028-06-30"},
	//     "IssuingCountry": {"word": "CHN"},
	//     "Number": {"word": "E12345678"}
	//   }
	// }

	// Placeholder implementation - in production, parse the actual API response
	return &model.OCRResult{
		Success:      false,
		ErrorMessage: "OCR adapter not fully configured - implement with valid Baidu API credentials",
	}, nil
}

// MockOCRAdapter is a mock implementation for testing.
type MockOCRAdapter struct {
	Result *model.OCRResult
	Error  error
}

// NewMockOCRAdapter creates a new MockOCRAdapter with predefined results.
func NewMockOCRAdapter(result *model.OCRResult, err error) *MockOCRAdapter {
	return &MockOCRAdapter{Result: result, Error: err}
}

// RecognizePassport returns the mock result.
func (m *MockOCRAdapter) RecognizePassport(ctx context.Context, imageReader io.Reader) (*model.OCRResult, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Result, nil
}
