package handlers

import (
	"AssetManagement/db"
	"AssetManagement/db/dbHelper"
	"AssetManagement/middleware"
	"AssetManagement/models"
	"AssetManagement/utils"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateAsset(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.UserContext(r)
	if userCtx == nil {
		utils.RespondError(w, nil, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body models.AssetRequest

	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, err, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(body); err != nil {
		utils.RespondError(w, err, "validation failed", http.StatusBadRequest)
		return
	}

	err := db.Tx(func(tx *sqlx.Tx) error {

		assetID, err := dbHelper.CreateAssetTx(tx, body)
		if err != nil {
			return err
		}

		switch body.AssetType {

		case "laptop":
			if body.Laptop == nil {
				return errors.New("laptop details missing")
			}
			return dbHelper.CreateLaptopTx(tx, assetID, body.Laptop)

		case "keyboard":
			if body.Keyboard == nil {
				return errors.New("keyboard details missing")
			}
			return dbHelper.CreateKeyboardTx(tx, assetID, body.Keyboard)

		case "mouse":
			if body.Mouse == nil {
				return errors.New("mouse details missing")
			}
			return dbHelper.CreateMouseTx(tx, assetID, body.Mouse)

		case "mobile":
			if body.Mobile == nil {
				return errors.New("mobile details missing")
			}
			return dbHelper.CreateMobileTx(tx, assetID, body.Mobile)
		}

		return nil
	})

	if err != nil {
		utils.RespondError(w, err, "asset creation failed", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusCreated, map[string]string{
		"message": "asset created successfully",
	})
}

func UpdateAsset(w http.ResponseWriter, r *http.Request) {
	assetID := chi.URLParam(r, "asset_id")
	if assetID == "" {
		utils.RespondError(w, nil, "invalid id", http.StatusBadRequest)
		return
	}

	var body models.UpdateAssetRequest

	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, err, "invalid body", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(body); err != nil {
		utils.RespondError(w, err, "validation failed", http.StatusBadRequest)
		return
	}

	if body.WarrantyExpiryDate.Before(body.WarrantyStart) {
		utils.RespondError(w, nil, "invalid warranty dates is given", http.StatusBadRequest)
		return
	}

	err := db.Tx(func(tx *sqlx.Tx) error {

		err := dbHelper.UpdateAssetTx(
			tx,
			assetID,
			body.Brand,
			body.Model,
			body.SerialNo,
			body.AssetType,
			body.Owner,
			body.WarrantyStart,
			body.WarrantyExpiryDate,
		)
		if err != nil {
			return err
		}

		switch body.AssetType {

		case "laptop":
			if body.Laptop == nil {
				return errors.New("laptop details required")
			}
			return dbHelper.UpdateLaptop(tx, assetID, body.Laptop)

		case "mouse":
			if body.Mouse == nil {
				return errors.New("mouse details required")
			}
			return dbHelper.UpdateMouse(tx, assetID, body.Mouse)

		case "keyboard":
			if body.Keyboard == nil {
				return errors.New("keyboard details required")
			}
			return dbHelper.UpdateKeyboard(tx, assetID, body.Keyboard)

		case "mobile":
			if body.Mobile == nil {
				return errors.New("mobile details arerequired")
			}
			return dbHelper.UpdateMobile(tx, assetID, body.Mobile)

		default:
			return errors.New("not a valid asset type")
		}
	})

	if err != nil {
		utils.RespondError(w, err, "failed to update asset", http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "asset updated successfully",
	})
}

func DeleteAsset(w http.ResponseWriter, r *http.Request) {
	assetID := chi.URLParam(r, "asset_id")

	err := dbHelper.DeleteAsset(assetID)
	if err != nil {
		utils.RespondError(w, err, "failed to delete asset", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "asset deleted successfully",
	})
}

func GetAssetsDashboard(w http.ResponseWriter, r *http.Request) {

	status := r.URL.Query().Get("status")
	assetType := r.URL.Query().Get("type")
	brand := r.URL.Query().Get("brand")
	owner := r.URL.Query().Get("owner")

	assets := make([]models.AssetList, 0)
	var err error

	if status != "" || assetType != "" || brand != "" || owner != "" {
		assets, err = dbHelper.GetAssetsWithFilters(status, assetType, brand, owner)
	} else {
		assets, err = dbHelper.GetAllAssetsForDashboard()

	}
	if err != nil {
		utils.RespondError(w, err, "failed to fetch assets", http.StatusInternalServerError)
		return
	}

	summary, err := dbHelper.GetAssetSummary()
	if err != nil {
		utils.RespondError(w, err, "failed to fetch summary", http.StatusInternalServerError)
		return
	}
	utils.RespondJSON(w, http.StatusOK, models.GetAllAssetsResponse{
		Summary: summary,
		Assets:  assets,
	})
}

func AssignAsset(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.UserContext(r)
	if userCtx == nil {
		utils.RespondError(w, nil, "unauthorized", http.StatusUnauthorized)
		return
	}

	assetID := chi.URLParam(r, "asset_id")
	if assetID == "" {
		utils.RespondError(w, nil, "asset id is required", http.StatusBadRequest)
		return
	}

	var body models.AssignAssetRequest
	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, err, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(body); err != nil {
		utils.RespondError(w, err, "validation failed", http.StatusBadRequest)
		return
	}

	if body.EmployeeID == uuid.Nil {
		utils.RespondError(w, nil, "employee_id is required", http.StatusBadRequest)
		return
	}

	err := db.Tx(func(tx *sqlx.Tx) error {
		return dbHelper.AssignAssetTx(
			tx,
			assetID,
			body.EmployeeID.String(),
			userCtx.UserID,
		)
	})

	if err != nil {
		utils.RespondError(w, err, "failed to assign asset", http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "asset assigned successfully",
	})
}

func ReturnAsset(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.UserContext(r)
	if userCtx == nil {
		utils.RespondError(w, nil, "unauthorized", http.StatusUnauthorized)
		return
	}

	assetID := chi.URLParam(r, "asset_id")
	if assetID == "" {
		utils.RespondError(w, nil, "asset id required", http.StatusBadRequest)
		return
	}

	err := db.Tx(func(tx *sqlx.Tx) error {
		return dbHelper.ReturnAssetTx(tx, assetID)
	})

	if err != nil {
		utils.RespondError(w, err, "failed to return asset", http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "asset returned successfully",
	})
}

func GetAssetByID(w http.ResponseWriter, r *http.Request) {
	assetID := chi.URLParam(r, "asset_id")

	asset, err := dbHelper.GetAssetByID(assetID)
	if err != nil {
		utils.RespondError(w, err, err.Error(), http.StatusNotFound)
		return
	}

	utils.RespondJSON(w, http.StatusOK, asset)
}

func SentToService(w http.ResponseWriter, r *http.Request) {

	userCtx := middleware.UserContext(r)
	if userCtx == nil {
		utils.RespondError(w, nil, "unauthorized", http.StatusUnauthorized)
		return
	}

	assetID := chi.URLParam(r, "asset_id")
	if assetID == "" {
		utils.RespondError(w, nil, "invalid asset id", http.StatusBadRequest)
		return
	}

	var body models.ServiceRequest
	//validate
	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, err, "invalid body", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(body); err != nil {
		utils.RespondError(w, err, "validation failed", http.StatusBadRequest)
		return
	}
	// parse time
	serviceStart, err := time.Parse("2006-01-02", body.StartDate)
	if err != nil {
		utils.RespondError(w, nil, "invalid start date format ", http.StatusBadRequest)
		return
	}

	serviceEnd, err := time.Parse("2006-01-02", body.EndDate)
	if err != nil {
		utils.RespondError(w, nil, "invalid end date format ", http.StatusBadRequest)
		return
	}

	if serviceEnd.Before(serviceStart) {
		utils.RespondError(w, nil, "End date must be after Start date", http.StatusBadRequest)
		return
	}

	err = db.Tx(func(tx *sqlx.Tx) error {
		return dbHelper.SentToServiceTx(
			tx,
			assetID,
			serviceStart,
			serviceEnd,
			userCtx.UserID,
		)
	})

	if err != nil {
		utils.RespondError(w, err, "failed to send to service", http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "asset is sent to service ",
	})
}
