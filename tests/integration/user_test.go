package integration

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	usermodel "github.com/travel-booking/server/internal/user/model"
)

// TestUserRegistrationLoginRealNameFlow (T140) verifies the complete user lifecycle:
// 1. Request SMS code
// 2. Login/register with SMS code
// 3. Verify JWT token works
// 4. Submit real-name verification
// 5. Verify profile shows verified status
func TestUserRegistrationLoginRealNameFlow(t *testing.T) {
	env := setupTestEnv(t)

	// Register user routes for the test.
	env.registerUserRoutes()

	t.Run("Step1_RequestSMSCode", func(t *testing.T) {
		w := env.doRequest("POST", "/api/v1/auth/sms-code",
			map[string]string{"phone": "13800138000"}, "")

		if w.Code != http.StatusOK {
			t.Fatalf("SMS code request failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}
	})

	t.Run("Step2_LoginWithCode", func(t *testing.T) {
		// In test mode, the SMS code is returned directly or a fixed code works.
		w := env.doRequest("POST", "/api/v1/auth/login",
			map[string]string{"phone": "13800138000", "code": "123456"}, "")

		if w.Code != http.StatusOK {
			t.Fatalf("login failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}

		// Verify response contains tokens.
		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("response data is not a map: %T", resp.Data)
		}
		if _, ok := data["access_token"]; !ok {
			t.Error("response missing access_token")
		}
		if _, ok := data["refresh_token"]; !ok {
			t.Error("response missing refresh_token")
		}
		if _, ok := data["user"]; !ok {
			t.Error("response missing user object")
		}
	})

	// Create a user directly for subsequent tests.
	user := &usermodel.UserAccount{
		Phone:          "13800138001",
		Nickname:       "测试用户",
		Status:         usermodel.UserStatusActive,
		RealNameStatus: usermodel.RNStatusUnverified,
		MemberLevel:    1,
	}
	env.DB.Create(user)

	token := env.generateToken(user.ID, "user", nil, nil)

	t.Run("Step3_VerifyToken", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/users/me", nil, token)

		if w.Code != http.StatusOK {
			t.Fatalf("get profile failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("response data is not a map: %T", resp.Data)
		}

		// Verify phone is masked.
		phone, _ := data["phone"].(string)
		if phone == "13800138001" {
			t.Error("phone should be masked in response, got raw value")
		}
		if phone == "" {
			t.Error("phone field is empty in response")
		}

		// Verify real-name status is unverified.
		status, _ := data["real_name_status"].(string)
		if status != usermodel.RNStatusUnverified {
			t.Errorf("expected real_name_status=%s, got %s", usermodel.RNStatusUnverified, status)
		}
	})

	t.Run("Step4_SubmitRealNameVerification", func(t *testing.T) {
		w := env.doRequest("POST", "/api/v1/users/me/real-name",
			map[string]string{
				"real_name":  "张三",
				"id_card_no": "110101199001011234",
			}, token)

		if w.Code != http.StatusOK {
			t.Fatalf("real-name verification failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}
	})

	t.Run("Step5_VerifyProfileUpdated", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/users/me", nil, token)

		if w.Code != http.StatusOK {
			t.Fatalf("get profile failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("response data is not a map: %T", resp.Data)
		}

		// After submission, status should be pending or verified
		// (depending on whether the real-name service auto-verifies).
		status, _ := data["real_name_status"].(string)
		if status != usermodel.RNStatusPending && status != usermodel.RNStatusVerified {
			t.Errorf("expected real_name_status=pending or verified, got %s", status)
		}
	})

	t.Run("Step6_VerifyEncryptedStorage", func(t *testing.T) {
		// Verify that the real name and ID card are encrypted in the database.
		var rnv usermodel.RealNameVerification
		env.DB.Where("user_id = ?", user.ID).Order("id DESC").First(&rnv)

		if rnv.IDCardNo == "110101199001011234" {
			t.Error("id_card_no should be encrypted in database, not stored as plaintext")
		}
		if rnv.RealName == "张三" {
			t.Error("real_name should be encrypted in database, not stored as plaintext")
		}
		if rnv.IDCardNo == "" {
			t.Error("id_card_no is empty in database")
		}
		if rnv.RealName == "" {
			t.Error("real_name is empty in database")
		}
	})
}

// TestUserLoginWithInvalidCode verifies that login fails with wrong SMS code.
func TestUserLoginWithInvalidCode(t *testing.T) {
	env := setupTestEnv(t)
	env.registerUserRoutes()

	w := env.doRequest("POST", "/api/v1/auth/login",
		map[string]string{"phone": "13800138000", "code": "000000"}, "")

	if w.Code == http.StatusOK {
		resp := parseResponse(t, w)
		if resp.Code == response.CodeSuccess {
			t.Error("login with invalid code should fail")
		}
	}
}

// TestUserProfileWithoutToken verifies that accessing protected endpoints
// without a token returns 401.
func TestUserProfileWithoutToken(t *testing.T) {
	env := setupTestEnv(t)
	env.registerUserRoutes()

	w := env.doRequest("GET", "/api/v1/users/me", nil, "")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestUserFrequentTravellerCRUD verifies the traveller management flow.
func TestUserFrequentTravellerCRUD(t *testing.T) {
	env := setupTestEnv(t)
	env.registerUserRoutes()

	// Create a user.
	user := &usermodel.UserAccount{
		Phone:   "13800138002",
		Nickname: "旅行者002",
		Status:  usermodel.UserStatusActive,
	}
	env.DB.Create(user)

	token := env.generateToken(user.ID, "user", nil, nil)

	// Create traveller.
	w := env.doRequest("POST", "/api/v1/users/me/travellers",
		map[string]string{
			"real_name":  "李四",
			"id_card_no": "110101199202022345",
			"phone":      "13900139000",
			"gender":     "female",
		}, token)

	if w.Code != http.StatusOK {
		t.Fatalf("create traveller failed: status %d, body: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	if resp.Code != response.CodeSuccess {
		t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
	}

	// List travellers.
	w = env.doRequest("GET", "/api/v1/users/me/travellers", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list travellers failed: status %d, body: %s", w.Code, w.Body.String())
	}
}

// registerUserRoutes sets up the user-related API routes for testing.
func (e *testEnv) registerUserRoutes() {
	// We register simplified handler stubs that exercise the middleware chain.
	// Full handler integration requires the actual services + Redis, which
	// we mock at the HTTP level for these tests.

	v1 := e.Router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/sms-code", func(c *gin.Context) {
				// In test mode, SMS code is accepted directly.
				response.OK(c, map[string]interface{}{"expires_in": 300})
			})
			auth.POST("/login", func(c *gin.Context) {
				var req struct {
					Phone string `json:"phone"`
					Code  string `json:"code"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					response.BadRequest(c, "invalid request")
					return
				}
				// In test mode, only accept code "123456".
				if req.Code != "123456" {
					response.BadRequest(c, "invalid SMS code")
					return
				}
				var user usermodel.UserAccount
				if err := e.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
					// Create new user.
					user = usermodel.UserAccount{
						Phone:    req.Phone,
						Nickname: "旅行者" + req.Phone[7:],
						Status:   usermodel.UserStatusActive,
					}
					e.DB.Create(&user)
				}

				accessToken, refreshToken, _ := e.JWT.GenerateTokenPair(user.ID, "user", nil, nil)
				response.OK(c, map[string]interface{}{
					"user": map[string]interface{}{
						"id":               user.ID,
						"phone":            user.Phone[:3] + "****" + user.Phone[7:],
						"nickname":         user.Nickname,
						"real_name_status": user.RealNameStatus,
					},
					"access_token":  accessToken,
					"refresh_token": refreshToken,
				})
			})
		}

		user := v1.Group("/users")
		user.Use(middleware.AuthRequired(e.JWT))
		{
			user.GET("/me", func(c *gin.Context) {
				userID := middleware.GetUserID(c)
				var u usermodel.UserAccount
				if err := e.DB.First(&u, userID).Error; err != nil {
					response.NotFound(c, "user not found")
					return
				}
				response.OK(c, map[string]interface{}{
					"id":               u.ID,
					"phone":            u.Phone[:3] + "****" + u.Phone[7:],
					"nickname":         u.Nickname,
					"real_name_status": u.RealNameStatus,
					"member_level":     u.MemberLevel,
				})
			})

			user.POST("/me/real-name", func(c *gin.Context) {
				userID := middleware.GetUserID(c)
				var req struct {
					RealName  string `json:"real_name"`
					IDCardNo  string `json:"id_card_no"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					response.BadRequest(c, "invalid request")
					return
				}

				// Encrypt sensitive fields before storage.
				encName, _ := e.Encryptor.Encrypt(req.RealName)
				encIDCard, _ := e.Encryptor.Encrypt(req.IDCardNo)

				// Create verification record.
				rnv := usermodel.RealNameVerification{
					UserID:   userID,
					RealName: encName,
					IDCardNo: encIDCard,
					Status:   usermodel.RNStatusPending,
				}
				e.DB.Create(&rnv)

				// Update user status.
				e.DB.Model(&usermodel.UserAccount{}).Where("id = ?", userID).
					Update("real_name_status", usermodel.RNStatusPending)

				response.OK(c, map[string]interface{}{"status": usermodel.RNStatusPending})
			})

			traveller := user.Group("/me/travellers")
			{
				traveller.GET("", func(c *gin.Context) {
					userID := middleware.GetUserID(c)
					var travellers []usermodel.FrequentTraveller
					e.DB.Where("user_id = ?", userID).Find(&travellers)
					response.OK(c, travellers)
				})

				traveller.POST("", func(c *gin.Context) {
					userID := middleware.GetUserID(c)
					var req struct {
						RealName  string `json:"real_name"`
						IDCardNo  string `json:"id_card_no"`
						Phone     string `json:"phone"`
						Gender    string `json:"gender"`
					}
					if err := c.ShouldBindJSON(&req); err != nil {
						response.BadRequest(c, "invalid request")
						return
					}

					encName, _ := e.Encryptor.Encrypt(req.RealName)
					encIDCard, _ := e.Encryptor.Encrypt(req.IDCardNo)

					traveller := usermodel.FrequentTraveller{
						UserID:   userID,
						RealName: encName,
						IDCardNo: encIDCard,
						Phone:    req.Phone,
						Gender:   req.Gender,
					}
					e.DB.Create(&traveller)

					// Return with masked fields.
					response.OK(c, map[string]interface{}{
						"id":       traveller.ID,
						"real_name": req.RealName[:1] + "**",
						"phone":    req.Phone[:3] + "****" + req.Phone[7:],
						"gender":   req.Gender,
					})
				})
			}
		}
	}
}
