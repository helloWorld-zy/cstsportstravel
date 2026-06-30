// Package service provides business logic for the Order domain.
package service

import (
	"fmt"

	"github.com/travel-booking/server/internal/product/model"
)

// VisaFormField represents a single field in a visa application form.
type VisaFormField struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Type        string   `json:"type"` // text, date, select, file, number, textarea
	Required    bool     `json:"required"`
	Placeholder string   `json:"placeholder,omitempty"`
	Options     []string `json:"options,omitempty"` // for select type
	MaxLength   int      `json:"max_length,omitempty"`
	Pattern     string   `json:"pattern,omitempty"` // regex pattern for validation
	Group       string   `json:"group,omitempty"`   // field group for layout
}

// VisaFormTemplate represents a complete visa application form template.
type VisaFormTemplate struct {
	CountryID   int64            `json:"country_id"`
	CountryName string           `json:"country_name"`
	VisaType    string           `json:"visa_type"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Fields      []VisaFormField  `json:"fields"`
	Groups      []VisaFieldGroup `json:"groups"`
}

// VisaFieldGroup represents a group of fields in the form.
type VisaFieldGroup struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// VisaFormService generates visa application form fields dynamically.
type VisaFormService struct{}

// NewVisaFormService creates a new VisaFormService.
func NewVisaFormService() *VisaFormService {
	return &VisaFormService{}
}

// GenerateForm generates a visa application form based on country and visa type.
func (s *VisaFormService) GenerateForm(country *model.Country) *VisaFormTemplate {
	template := &VisaFormTemplate{
		CountryID:   country.ID,
		CountryName: country.NameCN,
		VisaType:    country.VisaType,
		Title:       fmt.Sprintf("%s签证申请表", country.NameCN),
		Description: fmt.Sprintf("请填写%s签证申请所需的个人信息", country.NameCN),
		Groups: []VisaFieldGroup{
			{Name: "personal", Label: "个人信息", Description: "请填写与护照一致的个人信息"},
			{Name: "passport", Label: "护照信息", Description: "请填写护照详细信息"},
			{Name: "travel", Label: "旅行信息", Description: "请填写本次旅行相关信息"},
			{Name: "employment", Label: "职业信息", Description: "请填写当前职业状况"},
			{Name: "contact", Label: "联系方式", Description: "请填写紧急联系人信息"},
		},
	}

	// Common fields for all countries
	template.Fields = append(template.Fields, s.getPersonalFields()...)
	template.Fields = append(template.Fields, s.getPassportFields()...)
	template.Fields = append(template.Fields, s.getTravelFields()...)
	template.Fields = append(template.Fields, s.getEmploymentFields()...)
	template.Fields = append(template.Fields, s.getContactFields()...)

	// Add country-specific fields
	template.Fields = append(template.Fields, s.getCountrySpecificFields(country)...)

	return template
}

// getPersonalFields returns common personal information fields.
func (s *VisaFormService) getPersonalFields() []VisaFormField {
	return []VisaFormField{
		{Name: "surname_cn", Label: "姓（中文）", Type: "text", Required: true, Group: "personal", MaxLength: 50},
		{Name: "given_name_cn", Label: "名（中文）", Type: "text", Required: true, Group: "personal", MaxLength: 50},
		{Name: "surname_en", Label: "姓（拼音/英文）", Type: "text", Required: true, Group: "personal", MaxLength: 50, Placeholder: "与护照一致"},
		{Name: "given_name_en", Label: "名（拼音/英文）", Type: "text", Required: true, Group: "personal", MaxLength: 50, Placeholder: "与护照一致"},
		{Name: "gender", Label: "性别", Type: "select", Required: true, Group: "personal", Options: []string{"男", "女"}},
		{Name: "birth_date", Label: "出生日期", Type: "date", Required: true, Group: "personal"},
		{Name: "birth_place", Label: "出生地", Type: "text", Required: true, Group: "personal", MaxLength: 100},
		{Name: "nationality", Label: "国籍", Type: "text", Required: true, Group: "personal", MaxLength: 50},
		{Name: "marital_status", Label: "婚姻状况", Type: "select", Required: false, Group: "personal", Options: []string{"未婚", "已婚", "离异", "丧偶"}},
		{Name: "id_card_number", Label: "身份证号码", Type: "text", Required: true, Group: "personal", Pattern: `^\d{17}[\dX]$`},
	}
}

// getPassportFields returns passport information fields.
func (s *VisaFormService) getPassportFields() []VisaFormField {
	return []VisaFormField{
		{Name: "passport_number", Label: "护照号码", Type: "text", Required: true, Group: "passport", MaxLength: 20},
		{Name: "passport_type", Label: "护照类型", Type: "select", Required: true, Group: "passport", Options: []string{"普通护照", "公务护照", "外交护照"}},
		{Name: "passport_issue_date", Label: "签发日期", Type: "date", Required: true, Group: "passport"},
		{Name: "passport_expiry_date", Label: "有效期至", Type: "date", Required: true, Group: "passport"},
		{Name: "passport_issue_place", Label: "签发地", Type: "text", Required: true, Group: "passport", MaxLength: 100},
		{Name: "passport_photo", Label: "护照照片页", Type: "file", Required: true, Group: "passport"},
	}
}

// getTravelFields returns travel information fields.
func (s *VisaFormService) getTravelFields() []VisaFormField {
	return []VisaFormField{
		{Name: "purpose", Label: "出行目的", Type: "select", Required: true, Group: "travel", Options: []string{"旅游", "商务", "探亲", "留学"}},
		{Name: "entry_date", Label: "预计入境日期", Type: "date", Required: true, Group: "travel"},
		{Name: "exit_date", Label: "预计离境日期", Type: "date", Required: true, Group: "travel"},
		{Name: "entry_port", Label: "入境口岸", Type: "text", Required: false, Group: "travel", MaxLength: 100},
		{Name: "hotel_name", Label: "入住酒店", Type: "text", Required: false, Group: "travel", MaxLength: 200},
		{Name: "hotel_address", Label: "酒店地址", Type: "text", Required: false, Group: "travel", MaxLength: 500},
		{Name: "local_contact", Label: "当地联系人", Type: "text", Required: false, Group: "travel", MaxLength: 100},
		{Name: "local_contact_phone", Label: "当地联系人电话", Type: "text", Required: false, Group: "travel", MaxLength: 20},
	}
}

// getEmploymentFields returns employment information fields.
func (s *VisaFormService) getEmploymentFields() []VisaFormField {
	return []VisaFormField{
		{Name: "occupation_type", Label: "职业类型", Type: "select", Required: true, Group: "employment", Options: []string{"在职", "自由职业", "退休", "学生", "儿童"}},
		{Name: "company_name", Label: "工作单位", Type: "text", Required: false, Group: "employment", MaxLength: 200},
		{Name: "company_address", Label: "单位地址", Type: "text", Required: false, Group: "employment", MaxLength: 500},
		{Name: "company_phone", Label: "单位电话", Type: "text", Required: false, Group: "employment", MaxLength: 20},
		{Name: "position", Label: "职务", Type: "text", Required: false, Group: "employment", MaxLength: 100},
		{Name: "monthly_income", Label: "月收入（元）", Type: "number", Required: false, Group: "employment"},
		{Name: "school_name", Label: "学校名称", Type: "text", Required: false, Group: "employment", MaxLength: 200},
	}
}

// getContactFields returns emergency contact fields.
func (s *VisaFormService) getContactFields() []VisaFormField {
	return []VisaFormField{
		{Name: "emergency_contact_name", Label: "紧急联系人", Type: "text", Required: true, Group: "contact", MaxLength: 100},
		{Name: "emergency_contact_phone", Label: "紧急联系人电话", Type: "text", Required: true, Group: "contact", MaxLength: 20},
		{Name: "emergency_contact_relation", Label: "与申请人关系", Type: "select", Required: true, Group: "contact", Options: []string{"父母", "配偶", "子女", "兄弟姐妹", "朋友", "其他"}},
		{Name: "home_address", Label: "家庭住址", Type: "textarea", Required: true, Group: "contact", MaxLength: 500},
		{Name: "phone", Label: "手机号码", Type: "text", Required: true, Group: "contact", MaxLength: 20},
		{Name: "email", Label: "电子邮箱", Type: "text", Required: false, Group: "contact", MaxLength: 100},
	}
}

// getCountrySpecificFields returns additional fields specific to a country.
func (s *VisaFormService) getCountrySpecificFields(country *model.Country) []VisaFormField {
	var fields []VisaFormField

	switch country.VisaType {
	case model.VisaTypeRequired:
		// Countries requiring visa may need additional fields
		fields = append(fields, VisaFormField{
			Name:     "previous_visa",
			Label:    "是否曾获得该国签证",
			Type:     "select",
			Required: false,
			Group:    "travel",
			Options:  []string{"是", "否"},
		})
		fields = append(fields, VisaFormField{
			Name:     "previous_visa_number",
			Label:    " previous签证号码",
			Type:     "text",
			Required: false,
			Group:    "travel",
			MaxLength: 20,
		})

		// Schengen countries need additional fields
		if s.isSchengenCountry(country.NameEN) {
			fields = append(fields, VisaFormField{
				Name:     "schengen_main_destination",
				Label:    "申根主要目的地",
				Type:     "text",
				Required: true,
				Group:    "travel",
				MaxLength: 100,
			})
			fields = append(fields, VisaFormField{
				Name:     "schengen_first_entry",
				Label:    "申根首次入境国",
				Type:     "text",
				Required: true,
				Group:    "travel",
				MaxLength: 100,
			})
			fields = append(fields, VisaFormField{
				Name:     "insurance_policy_number",
				Label:    "申根保险保单号",
				Type:     "text",
				Required: true,
				Group:    "travel",
				MaxLength: 50,
			})
		}
	}

	return fields
}

// isSchengenCountry checks if a country is a Schengen member.
func (s *VisaFormService) isSchengenCountry(countryEN string) bool {
	schengenCountries := map[string]bool{
		"Austria": true, "Belgium": true, "Czech Republic": true,
		"Denmark": true, "Estonia": true, "Finland": true,
		"France": true, "Germany": true, "Greece": true,
		"Hungary": true, "Iceland": true, "Italy": true,
		"Latvia": true, "Liechtenstein": true, "Lithuania": true,
		"Luxembourg": true, "Malta": true, "Netherlands": true,
		"Norway": true, "Poland": true, "Portugal": true,
		"Slovakia": true, "Slovenia": true, "Spain": true,
		"Sweden": true, "Switzerland": true,
	}
	return schengenCountries[countryEN]
}
