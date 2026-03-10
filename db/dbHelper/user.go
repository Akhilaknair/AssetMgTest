package dbHelper

import (
	"AssetManagement/db"
	"AssetManagement/models"
	"AssetManagement/utils"
	"time"

	"github.com/jmoiron/sqlx"
)

func GetUserID(body models.LoginRequest) (string, error) {
	query := `select id , password 
                from users 
                where email=trim($1)
                and archived_at is null`

	var user models.LoginData
	if err := db.Assets.Get(&user, query, body.Email); err != nil {
		return "", err
	}
	if passwordErr := utils.CheckPassword(body.Password, user.PasswordHash); passwordErr != nil {
		return "", passwordErr
	}
	return user.ID, nil
}

//	func GetEmployeeRole(userID string) (string, error) {
//		query := `select role from users
//	             where id = $1 and archived_at is null`
//
//		var role string
//		err := db.Assets.Get(&role, query, userID)
//
//		if err != nil {
//			if err == sql.ErrNoRows {
//				return "", err
//			}
//			return "", err
//		}
//
//		return role, nil
//	}
func IsUserExists(email string) (bool, error) {
	query := `select count(id)>0 as is_exists
                 from users
                 where email =trim($1)
                 and archived_at is null`
	var check bool
	err := db.Assets.Get(&check, query, email)
	return check, err
}

func CreateSession(userID string) (string, error) {
	var sessionID string

	query := `insert into user_session(user_id) 
             values($1) returning id`

	err := db.Assets.Get(&sessionID, query, userID)
	return sessionID, err
}

func DeleteSession(sessionID string) error {
	query := `update user_session
               set archived_at = now() 
               where id=$1
               and archived_at is null`

	_, err := db.Assets.Exec(query, sessionID)
	return err
}

func GetUserById(userID string) (models.User, error) {
	var user models.User
	query := `select id,name,email,phone_no,user_type,role
              from users 
              where id=$1
              and archived_at is null`

	err := db.Assets.Get(&user, query, userID)
	return user, err
}

func GetArchivedAt(sessionID string) (*time.Time, error) {
	var archivedAt *time.Time
	query := `select archived_at
              from user_session
              where id=$1`

	getErr := db.Assets.Get(&archivedAt, query, sessionID)
	return archivedAt, getErr
}

//func DeleteSessionTx(tx *sqlx.Tx, sessionID string) error {
//	query := `update user_session
//               set archived_at = now()
//               where id=$1
//               and archived_at is null`
//
//	_, err := tx.Exec(query, sessionID)
//	return err
//}
//
//func DeleteUserTx(tx *sqlx.Tx, userID string) error {
//	query := `update users
//	         set archived_at =now()
//	         where id=$1
//	         and archived_at is null`
//
//	_, delErr := tx.Exec(query, userID)
//	return delErr
//
//}

func UpdateUserRole(userID, role string) error {
	query := `
	update users
	set role = $1,updated_at = NOW()
	where id = $2 and archived_at is null`

	_, err := db.Assets.Exec(query, role, userID)
	return err
}
func CreateUserTx(tx *sqlx.Tx, name, email, password, phoneNo, userType string, joiningDate time.Time) (string, error) {

	query := `insert into users(name , email, password ,phone_no , role, user_type,joining_date)
               values (trim($1), trim($2), trim($3), trim($4), 'employee', $5,$6)
               returning id`

	var userID string
	err := tx.Get(&userID, query, name, email, password, phoneNo, userType, joiningDate)
	return userID, err
}

func CreateSessionTx(tx *sqlx.Tx, userID string) (string, error) {
	var sessionID string

	query := `insert into user_session(user_id)
              values ($1) returning id`

	err := tx.Get(&sessionID, query, userID)
	return sessionID, err
}

func GetAllUsersWithAssetCount() ([]models.UserWithAssets, error) {

	query := `
		select
			u.id,u.name,
			u.email,u.phone_no,
		u.user_type,u.role,
		count(a.id) as assigned_count
		from users u left join assets a 
		on a.current_assigned_to = u.id 
		and a.archived_at is null
		where u.archived_at is null
		group by u.id
		order by u.created_at desc`

	var users []models.UserWithAssets
	err := db.Assets.Select(&users, query)
	if err != nil {
		return nil, err
	}

	for i := range users {
		assetsQuery := `
			select id, brand, model
			from assets
			where current_assigned_to = $1
			and archived_at is null`

		assets := make([]models.AssignedAssetInfo, 0)
		err := db.Assets.Select(&assets, assetsQuery, users[i].ID)
		if err != nil {
			return nil, err
		}

		users[i].AssignedAssets = assets
	}

	return users, nil
}

func GetUsersWithFilters(role, name, userType string) ([]models.UserWithAssets, error) {

	query := `select u.id, u.name, u.email, u.phone_no,
		u.user_type, u.role,
		count(a.id) as assigned_count
		from users u
		left join assets a
		on a.current_assigned_to = u.id
		and a.archived_at is null
		where u.archived_at is null
	and ($1 = '' OR u.role = $1::user_role)
		and ($2 = '' OR lower(u.name) like lower('%' || $2 || '%'))
		and ($3 = '' OR u.user_type = $3::user_type)
		group by u.id
		order by u.created_at desc
	`

	var users []models.UserWithAssets

	err := db.Assets.Select(&users, query, role, name, userType)
	if err != nil {
		return nil, err
	}

	for i := range users {
		query := `select id, brand, model
			from assets
			where current_assigned_to = $1
			and archived_at is null`

		var assets []models.AssignedAssetInfo
		if err := db.Assets.Select(&assets, query, users[i].ID); err != nil {
			return nil, err
		}
		users[i].AssignedAssets = assets
	}

	return users, nil
}

// transaction db-helper functions
func DeleteSessionTx(tx *sqlx.Tx, userID string) error {
	query := `update user_session
              set archived_at=now()
              where user_id=$1
              and archived_at is null`

	_, err := tx.Exec(query, userID)
	return err
}

func DeleteUserTx(tx *sqlx.Tx, userID string) error {
	query := `update users
              set archived_at = now()
              where id=$1
              and archived_at is null`

	_, delErr := tx.Exec(query, userID)
	return delErr
}
