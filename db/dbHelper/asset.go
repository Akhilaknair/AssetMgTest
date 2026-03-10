package dbHelper

import (
	"AssetManagement/db"
	"AssetManagement/models"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

func CreateAssetTx(tx *sqlx.Tx, asset models.AssetRequest) (string, error) {

	var assetID string

	query := `
		insert into assets
			(brand, model, serial_no, asset_type, owner, warranty_start_date, warranty_expiry_date)
		values($1,$2,$3,$4,$5,$6,$7)
		returning id`

	err := tx.Get(
		&assetID,
		query,
		asset.Brand,
		asset.Model,
		asset.SerialNo,
		asset.AssetType,
		asset.Owner,
		asset.WarrantyStart,
		asset.WarrantyExpiryDate,
	)

	if err != nil {
		return "", err
	}

	return assetID, nil
}

func CreateLaptopTx(tx *sqlx.Tx, assetID string, req *models.LaptopRequest) error {
	query := `
		insert into laptop
			(asset_id, processor, ram, storage, os, charger, password)
		values($1,$2,$3,$4,$5,$6,$7)`

	_, err := tx.Exec(
		query,
		assetID,
		req.Processor,
		req.RAM,
		req.Storage,
		req.OS,
		req.Charger,
		req.Password,
	)
	return err
}

func CreateKeyboardTx(tx *sqlx.Tx, assetID string, req *models.KeyboardRequest) error {
	query := `
		insert into keyboard
			(asset_id, layout, connectivity)
		values ($1,$2,$3)`
	_, err := tx.Exec(query, assetID, req.Layout, req.Connectivity)
	return err
}

func CreateMobileTx(tx *sqlx.Tx, assetID string, req *models.MobileRequest) error {
	query := `
		insert into mobile
			(asset_id, os, ram, storage, charger, password)
		values ($1,$2,$3,$4,$5,$6)`

	_, err := tx.Exec(
		query,
		assetID,
		req.OS,
		req.RAM,
		req.Storage,
		req.Charger,
		req.Password,
	)
	return err
}

func CreateMouseTx(tx *sqlx.Tx, assetID string, req *models.MouseRequest) error {
	query := `
		insert into mouse
			(asset_id, dpi, connectivity)
		values($1,$2,$3)`

	_, err := tx.Exec(query, assetID, req.DPI, req.Connectivity)
	return err
}

func GetAssetByID(assetID string) (*models.Asset, error) {

	query := `select 
              id,brand,model,serial_no,asset_type , 
              status,owner,warranty_start_date,warranty_expiry_date
              from assets
              where id =$1 
              and archived_at is null`

	var asset models.Asset
	err := db.Assets.Get(&asset, query, assetID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("asset not found")
	}
	if err != nil {
		return nil, err
	}
	return &asset, nil

}

func GetAssetsByUserID(userID string) ([]models.AssignedAssetInfo, error) {
	query := ` select id, brand,model
		       from assets
		       where current_assigned_to = $1
		       and archived_at is null`

	assets := make([]models.AssignedAssetInfo, 0)
	err := db.Assets.Select(&assets, query, userID)
	return assets, err
}

func UpdateAssetTx(
	tx *sqlx.Tx,
	assetID string,
	brand, model, serialNo, assetType, owner string,
	warrantyStart, warrantyEnd time.Time,
) error {

	query := `
		update assets
		set brand = $2,
		    model = $3,
		    serial_no = $4,
		    asset_type = $5,
		    owner = $6,
		    warranty_start_date = $7,
		    warranty_expiry_date = $8,
		    updated_at = NOW()
		where id = $1
		and archived_at is null`

	_, err := tx.Exec(
		query,
		assetID,
		brand,
		model,
		serialNo,
		assetType,
		owner,
		warrantyStart,
		warrantyEnd,
	)

	return err
}

func DeleteAsset(assetID string) error {
	query := `
		update assets
		set archived_at = NOW()
		where id = $1 and archived_at is null`

	_, err := db.Assets.Exec(query, assetID)
	return err
}

func AssignAssetTx(
	tx *sqlx.Tx,
	assetID string,
	employeeID string,
	assignedBy string,
) error {

	var status string

	checkQuery := `Select status from assets
		           where id = $1
		           and archived_at is null`

	if err := tx.Get(&status, checkQuery, assetID); err != nil {
		return err
	}

	if status != "available" {
		return errors.New("asset is not available")
	}

	updateQuery := `
		update assets
		set status = 'assigned',
		    current_assigned_to = $1,
		    updated_at = NOW()
		where id = $2
		  and archived_at is null
	`
	if _, err := tx.Exec(updateQuery, employeeID, assetID); err != nil {
		return err
	}

	insertHistoryQuery := `
		insert into asset_history
		    (asset_id, assigned_to, assigned_by, assigned_on)
		values ($1, $2, $3, NOW())
	`
	_, err := tx.Exec(insertHistoryQuery, assetID, employeeID, assignedBy)

	return err
}

func ReturnAssetTx(tx *sqlx.Tx, assetID string) error {

	var assignedTo *string

	checkQuery := `
		select current_assigned_to
		from assets
		  where id = $1
		and archived_at is null`

	if err := tx.Get(&assignedTo, checkQuery, assetID); err != nil {
		return err
	}

	if assignedTo == nil {
		return errors.New("asset is not currently assigned")
	}

	updateAssetQuery := `
		update assets
	    set status = 'available',
		    current_assigned_to = null,
		    updated_at = NOW()
		where id = $1
		and archived_at is null
	`
	if _, err := tx.Exec(updateAssetQuery, assetID); err != nil {
		return err
	}

	updateHistoryQuery := `
		update asset_history
		set returned_on = NOW()
	     	where asset_id = $1
		and returned_on is null
	`
	_, err := tx.Exec(updateHistoryQuery, assetID)

	return err
}

func GetAssetSummary() (models.AssetSummary, error) {
	query := `
		select count(*) as total,
			count(*) filter (where status = 'available') as available,
			 count(*) filter (where status = 'assigned') as assigned,
		 count(*) filter (where status = 'waiting_for_repair') as waiting_for_repair,
			count(*) filter (where status = 'in_service') as service,
			 count(*) filter (where status = 'damaged') as damaged
		from assets
		where archived_at is null`

	var summary models.AssetSummary
	err := db.Assets.Get(&summary, query)
	return summary, err

}

func GetAllAssetsForDashboard() ([]models.AssetList, error) {

	query := `select 
		a.id,a.brand,
		a.model,a.asset_type,
		a.serial_no,a.status,
		u.name as assigned_to,
		a.owner as owned_by
		from assets a
		left join users u on a.current_assigned_to = u.id
		where a.archived_at is null
		order by a.created_at desc`

	assets := make([]models.AssetList, 0)
	err := db.Assets.Select(&assets, query)
	return assets, err
}

//func GetAssetsDashboard() (*models.GetAllAssetsResponse, error) {
//	summary, err := GetAssetSummary()
//	if err != nil {
//		return nil, err
//	}
//
//	assets, err := GetAllAssetsForDashboard()
//	if err != nil {
//		return nil, err
//	}
//
//	return &models.GetAllAssetsResponse{
//		Summary: summary,
//		Assets:  assets,
//	}, nil
//}

func UpdateLaptop(tx *sqlx.Tx, assetID string, req *models.LaptopRequest) error {

	query := `
		update laptop
		set processor = $2,
		 ram = $3,storage = $4,
		os = $5,
		charger = $6,password = $7
		where asset_id = $1`

	_, err := tx.Exec(
		query,
		assetID,
		req.Processor,
		req.RAM,
		req.Storage,
		req.OS,
		req.Charger,
		req.Password,
	)

	return err
}

func UpdateKeyboard(tx *sqlx.Tx, assetID string, req *models.KeyboardRequest) error {

	query := `update keyboard
		set layout = $2,connectivity = $3
		where asset_id = $1`

	_, err := tx.Exec(
		query,
		assetID,
		req.Layout,
		req.Connectivity,
	)
	return err
}

func UpdateMouse(tx *sqlx.Tx, assetID string, req *models.MouseRequest) error {

	query := `update mouse
		set dpi = $2,connectivity = $3
		where asset_id = $1`

	_, err := tx.Exec(
		query,
		assetID,
		req.DPI,
		req.Connectivity,
	)

	return err
}

func UpdateMobile(tx *sqlx.Tx, assetID string, req *models.MobileRequest) error {

	query := `update mobile
		set os = $2,ram = $3, storage = $4,
		charger = $5,password = $6
		where asset_id = $1`

	_, err := tx.Exec(
		query,
		assetID,
		req.OS,
		req.RAM,
		req.Storage,
		req.Charger,
		req.Password,
	)
	return err
}

func SentToServiceTx(tx *sqlx.Tx, assetID string, serviceStart, serviceEnd time.Time, userID string) error {
	//asset table update
	updateQuery := `
		update assets
		set status = 'in_service',
		    updated_at = now()
		where id = $1
		  and archived_at is null
		  and status = 'available'`

	res, err := tx.Exec(updateQuery, assetID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("asset not available ")
	}

	//asset history table updation
	insertQuery := `
		insert into asset_history (
			asset_id,
			assigned_to,
			assigned_by,
			assigned_on,
			service_start,
			service_end
		)
		values ($1, $2, $3, now(), $4, $5)
	`

	_, err = tx.Exec(insertQuery,
		assetID,
		userID,
		userID,
		serviceStart,
		serviceEnd,
	)

	return err
}

func GetAssetsWithFilters(status, assetType, brand, owner string) ([]models.AssetList, error) {

	query := `select a.id , a.brand , a.model ,a.asset_type,
                    a.serial_no,a.status ,u.name as assigned_to ,
                   a.owner as owned_by
                   from assets a left join users u 
                   on a.current_assigned_to =u.id
                   where a.archived_at is null
                   and ($1 ='' or a.status =$1::asset_status)
                   and ($2 ='' or a.asset_type =$2::asset_type)
                   and ($3 ='' or lower(a.brand) like lower('%' || $3 || '%'))
                   and ($4='' or a.owner=$4)
                   order by a.created_at desc`

	assets := make([]models.AssetList, 0)
	err := db.Assets.Select(&assets, query, status, assetType, brand, owner)
	return assets, err
}
