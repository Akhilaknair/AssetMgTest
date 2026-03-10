package handlers

import (
	"AssetManagement/db"
	"AssetManagement/db/dbHelper"
	"AssetManagement/middleware"
	"AssetManagement/models"
	"AssetManagement/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var body models.RegisterRequest

	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, parseErr, "parsing failed", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(body); err != nil {
		utils.RespondError(w, err, "input validation failed", http.StatusBadRequest)
		return
	}

	exists, err := dbHelper.IsUserExists(body.Email)
	if err != nil {
		utils.RespondError(w, err, "failed to check user existence", http.StatusInternalServerError)
		return
	}

	if exists {
		utils.RespondError(w, nil, "user already exist", http.StatusConflict)
		return
	}

	hashedPassword, hashErr := utils.HashPassword(body.Password)
	if hashErr != nil {
		utils.RespondError(w, hashErr, "failed to hash password", http.StatusInternalServerError)
		return
	}

	var userID, sessionID string
	var saveErr, crtErr error

	trxErr := db.Tx(func(tx *sqlx.Tx) error {
		userID, saveErr = dbHelper.CreateUserTx(
			tx,
			body.Name,
			body.Email,
			hashedPassword,
			body.PhoneNo,
			body.UserType,
			body.JoiningDate)

		if saveErr != nil {
			utils.RespondError(w, saveErr, "failed to create user", http.StatusInternalServerError)
			return saveErr
		}

		sessionID, crtErr = dbHelper.CreateSessionTx(tx, userID)
		if crtErr != nil {
			utils.RespondError(w, crtErr, "failed to create session", http.StatusInternalServerError)
			return crtErr
		}
		return nil
	})

	if trxErr != nil {
		utils.RespondError(w, trxErr, "user account creation failed ", http.StatusInternalServerError)
		return
	}

	user, getErr := dbHelper.GetUserById(userID)
	if getErr != nil {
		utils.RespondError(w, getErr, "failed to fetch user role", http.StatusInternalServerError)
		return
	}

	token, err := utils.GenerateJWT(sessionID, userID, user.Role)
	if err != nil {
		utils.RespondError(w, err, "failed to generate token", http.StatusInternalServerError)
		return
	}
	utils.RespondJSON(w, http.StatusCreated, struct {
		Message string `json:"message"`
		Token   string `json:"token"`
	}{"registered successfully & login successful", token})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var body models.LoginRequest
	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, parseErr, "invalid body , parsing failed ", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(body); err != nil {
		utils.RespondError(w, err, "input validation failed", http.StatusBadRequest)
		return
	}
	userID, userErr := dbHelper.GetUserID(body)
	if userErr != nil {
		utils.RespondError(w, nil, "invalid credentials ", http.StatusUnauthorized)
		return
	}
	user, err := dbHelper.GetUserById(userID)
	if err != nil {
		utils.RespondError(w, err, "failed to get user details", http.StatusInternalServerError)
		return
	}

	sessionID, crtErr := dbHelper.CreateSession(userID)
	if crtErr != nil {
		utils.RespondError(w, crtErr, "failed to create the session ", http.StatusInternalServerError)
		return
	}
	token, err := utils.GenerateJWT(sessionID, userID, user.Role)
	if err != nil {
		utils.RespondError(w, err, "failed to create the token ", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
		Token   string `json:"token"`
	}{"login successful", token})

}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.UserContext(r)

	if userCtx == nil || userCtx.SessionID == "" {
		utils.RespondError(w, nil, "unauthorized", http.StatusUnauthorized)
		return
	}
	sessionID := userCtx.SessionID
	if err := dbHelper.DeleteSession(sessionID); err != nil {
		utils.RespondError(w, err, "delete session failed", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{"logout successful"})

}

func GetUser(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.UserContext(r)

	if userCtx == nil || userCtx.UserID == "" {
		utils.RespondError(w, nil, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID := userCtx.UserID
	user, getErr := dbHelper.GetUserById(userID)

	if getErr != nil {
		utils.RespondError(w, getErr, "getting user detail failed", http.StatusInternalServerError)
		return
	}
	utils.RespondJSON(w, http.StatusOK, user)
}

func GetAssetsByEmpID(w http.ResponseWriter, r *http.Request) {

	user := middleware.UserContext(r)
	if user == nil {
		utils.RespondError(w, nil, "user id not found in context", http.StatusUnauthorized)
		return
	}

	assets, err := dbHelper.GetAssetsByUserID(user.UserID)
	if err != nil {
		utils.RespondError(w, err, "failed to fetch assets", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		TotalAssets int                        `json:"total_assets"`
		Assets      []models.AssignedAssetInfo `json:"assets"`
	}{
		TotalAssets: len(assets),
		Assets:      assets,
	})
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {

	//users, err := dbHelper.GetAllUsersWithAssetCount()

	roleFilter := r.URL.Query().Get("role")
	nameFilter := r.URL.Query().Get("name")
	typeFilter := r.URL.Query().Get("type")

	var users interface{}
	var err error

	if roleFilter != "" || nameFilter != "" || typeFilter != "" {
		users, err = dbHelper.GetUsersWithFilters(roleFilter, nameFilter, typeFilter)
	} else {
		users, err = dbHelper.GetAllUsersWithAssetCount()
	}

	if err != nil {
		utils.RespondError(w, err, "failed to fetch users", http.StatusInternalServerError)
		return
	}
	utils.RespondJSON(w, http.StatusOK, struct {
		Users interface{} `json:"users"`
	}{users})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.UserContext(r)

	if userCtx == nil || userCtx.UserID == "" {
		utils.RespondError(w, nil, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID := userCtx.UserID

	targetUserID := chi.URLParam(r, "user_id")
	if targetUserID == "" {
		utils.RespondError(w, nil, "user id is required", http.StatusBadRequest)
		return
	}

	if userID == targetUserID {
		utils.RespondError(w, nil, "cannot delete your own account", http.StatusForbidden)
		return
	}

	//	sessionID := userCtx.SessionID

	trxErr := db.Tx(func(tx *sqlx.Tx) error {
		if delErr := dbHelper.DeleteSessionTx(tx, targetUserID); delErr != nil {
			return delErr
		}
		return dbHelper.DeleteUserTx(tx, targetUserID)
	})

	if trxErr != nil {
		utils.RespondError(w, trxErr, "user account deletion failed", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{"account deleted successfully"})
}

func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	var body models.UpdateUserRoleRequest
	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, err, "invalid body", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(body); err != nil {
		utils.RespondError(w, err, "validation failed", http.StatusBadRequest)
		return
	}

	if err := dbHelper.UpdateUserRole(userID, body.Role); err != nil {
		utils.RespondError(w, err, "failed to update role", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "role updated successfully",
	})
}
