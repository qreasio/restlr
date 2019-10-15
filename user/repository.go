package user

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/qreasio/restlr/model"
	"github.com/qreasio/restlr/toolbox"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	GetUserByID(ctx context.Context, id uint64) (*model.UserDetail, error)
	GetUserByIDList(ctx context.Context, idList []uint64) (map[uint64]*model.UserDetail, error)
	UserDetailAsUserSlice(baseURL string, apiHost string, u *model.UserDetail) []*model.User
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (repo *repository) GetUserByID(ctx context.Context, id uint64) (*model.UserDetail, error) {
	apiConfig := ctx.Value(model.APICONFIGKEY).(model.APIModel)
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

func (repo *repository) GetUserByIDList(ctx context.Context, idList []uint64) (map[uint64]*model.UserDetail, error) {
	apiConfig := ctx.Value(model.APICONFIGKEY).(model.APIModel)
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

func (repo *repository) UserDetailAsUserSlice(baseURL string, apiHost string, u *model.UserDetail) []*model.User {
	var users []*model.User

	user := &u.User
	user.Link = apiHost + "/author/" + u.NiceName

	selfLink := baseURL + "/users/" + toolbox.UInt64ToStr(u.ID)
	if len(user.Links.SelfLink) < 1 {
		user.Links.SelfLink = append(user.Links.SelfLink, map[string]string{"href": selfLink})
	}

	collectionLink := baseURL + "/users"
	if len(user.Links.Collection) < 1 {
		user.Links.Collection = append(user.Links.Collection, map[string]string{"href": collectionLink})
	}

	users = append(users, user)
	return users
}
