package user

import (
	"context"
	"database/sql"

	"github.com/qreasio/restlr/model"
	"github.com/qreasio/restlr/toolbox"
	log "github.com/sirupsen/logrus"
)

// Repository is interface for functions to interact with database
type Repository interface {
	GetUserByID(ctx context.Context, id uint64) (*model.UserDetail, error)
	GetUserByIDList(ctx context.Context, idList []uint64) (map[uint64]*model.UserDetail, error)
}

type repository struct {
	db *sql.DB
}

// NewRepository is function to create new repository struct instance that implements Repository interface
func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

// GetUserByID is function to get UserDetail from specific id parameter
func (repo *repository) GetUserByID(ctx context.Context, id uint64) (*model.UserDetail, error) {
	apiConfig := ctx.Value(model.APIConfigKey).(model.APIConfig)
	tableName := apiConfig.TablePrefix + "users"
	metaTableName := apiConfig.TablePrefix + "usermeta"

	var sql = `SELECT ` +
		`u.ID, u.user_login, u.user_pass, u.user_nicename, u.user_email, u.user_url, u.user_registered, u.user_activation_key, u.user_status, u.display_name, m.meta_value ` +
		`FROM ` + tableName + ` u LEFT JOIN ` + metaTableName + ` m ON m.user_id = u.ID ` +
		`WHERE m.meta_key = 'description' AND u.ID = ?`

	wu := &model.UserDetail{}

	err := repo.db.QueryRow(sql, id).Scan(&wu.ID, &wu.Login, &wu.Pass, &wu.NiceName, &wu.Email, &wu.URL, &wu.Registered, &wu.ActivationKey, &wu.Status, &wu.DisplayName, &wu.Description)

	return wu, err
}

// GetUserByIDList is function to get list of UserDetail from idList parameter
func (repo *repository) GetUserByIDList(ctx context.Context, idList []uint64) (map[uint64]*model.UserDetail, error) {
	apiConfig := ctx.Value(model.APIConfigKey).(model.APIConfig)
	tableName := apiConfig.TablePrefix + "users"
	metaTableName := apiConfig.TablePrefix + "usermeta"

	var sqlQuery = `SELECT ` +
		`u.ID, u.user_login, u.user_pass, u.user_nicename, u.user_email, u.user_url, u.user_registered, u.user_activation_key, u.user_status, u.display_name, m.meta_value ` +
		`FROM ` + tableName + ` u LEFT JOIN ` + metaTableName + ` m ON m.user_id = u.ID ` +
		`WHERE m.meta_key = 'description' AND u.ID IN (` + toolbox.UInt64SliceToCSV(idList) + `)`

	q, err := repo.db.Query(sqlQuery)
	if err != nil {
		log.WithFields(log.Fields{
			"params": sqlQuery,
			"func":   "db.Query",
		}).Errorf("Failed to run db query: %s", err)
		return nil, err
	}
	defer q.Close()

	var res = make(map[uint64]*model.UserDetail)
	for q.Next() {
		wu := model.UserDetail{}
		err = q.Scan(&wu.ID, &wu.Login, &wu.Pass, &wu.NiceName, &wu.Email, &wu.URL, &wu.Registered, &wu.ActivationKey, &wu.Status, &wu.DisplayName, &wu.Description)
		if err != nil {
			log.WithFields(log.Fields{
				"params": idList,
				"func":   "q.Scan",
			}).Errorf("Failed to run query scan: %s", err)
			return nil, err
		}

		res[wu.ID] = &wu
	}

	return res, err
}
