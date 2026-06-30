package service

import (
	"context"
	"strings"
	"testing"

	"github.com/travel-booking/server/internal/order/model"
)

func TestMockOCRAdapter_Success(t *testing.T) {
	expected := &model.OCRResult{
		Name:           "张三",
		PassportNumber: "E12345678",
		ExpiryDate:     "2028-06-30",
		Nationality:    "中国",
		Success:        true,
	}

	adapter := NewMockOCRAdapter(expected, nil)
	result, err := adapter.RecognizePassport(context.Background(), strings.NewReader("test-image"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.Name != "张三" {
		t.Errorf("expected name '张三', got '%s'", result.Name)
	}
	if result.PassportNumber != "E12345678" {
		t.Errorf("expected passport number 'E12345678', got '%s'", result.PassportNumber)
	}
}

func TestMockOCRAdapter_Error(t *testing.T) {
	adapter := NewMockOCRAdapter(nil, context.DeadlineExceeded)
	_, err := adapter.RecognizePassport(context.Background(), strings.NewReader("test-image"))

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestMockOCRAdapter_FailedResult(t *testing.T) {
	expected := &model.OCRResult{
		Success:      false,
		ErrorMessage: "image too blurry",
	}

	adapter := NewMockOCRAdapter(expected, nil)
	result, err := adapter.RecognizePassport(context.Background(), strings.NewReader("test-image"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("expected success to be false")
	}
	if result.ErrorMessage != "image too blurry" {
		t.Errorf("expected error message 'image too blurry', got '%s'", result.ErrorMessage)
	}
}

func TestBaiduOCRAdapter_Creation(t *testing.T) {
	config := BaiduOCRConfig{
		APIKey:    "test-key",
		SecretKey: "test-secret",
	}
	adapter := NewBaiduOCRAdapter(config)

	if adapter == nil {
		t.Fatal("expected non-nil adapter")
	}
	if adapter.config.Endpoint != "https://aip.baidubce.com" {
		t.Errorf("expected default endpoint, got '%s'", adapter.config.Endpoint)
	}
}

func TestBaiduOCRAdapter_DefaultEndpoint(t *testing.T) {
	config := BaiduOCRConfig{
		APIKey:    "test-key",
		SecretKey: "test-secret",
	}
	adapter := NewBaiduOCRAdapter(config)

	if adapter.config.Endpoint == "" {
		t.Error("expected default endpoint to be set")
	}
}

func TestOCRResult_Fields(t *testing.T) {
	result := model.OCRResult{
		Name:           "John Doe",
		PassportNumber: "AB1234567",
		ExpiryDate:     "2030-01-15",
		Nationality:    "USA",
		Gender:         "M",
		BirthDate:      "1990-05-20",
		IssueDate:      "2020-01-15",
		IssuePlace:     "Washington DC",
		Success:        true,
	}

	if result.Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got '%s'", result.Name)
	}
	if result.Gender != "M" {
		t.Errorf("expected gender 'M', got '%s'", result.Gender)
	}
	if result.IssuePlace != "Washington DC" {
		t.Errorf("expected issue place 'Washington DC', got '%s'", result.IssuePlace)
	}
}
