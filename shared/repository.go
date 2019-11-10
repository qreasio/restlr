package shared

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/qreasio/restlr/model"
	"github.com/qreasio/restlr/toolbox"
	log "github.com/sirupsen/logrus"
)

// Repository is interface for functions that interact with database
type Repository interface {
	LoadOption(ctx context.Context, optionName string) (*model.Option, error)
	PostMetasByPostIDs(ctx context.Context, idList []uint64) (map[uint64]map[string]string, error)
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

// LoadOptions is function get Option list from list of option name
func (repo *repository) LoadOptions(ctx context.Context, autoload string, optionName []string) ([]*model.Option, error) {
	apiConfig := ctx.Value(model.APIConfigKey).(model.APIConfig)
	tableName := apiConfig.TablePrefix + "options"
	var err error

	var sqlQuery = `SELECT ` +
		`option_id, option_name, option_value, autoload ` +
		`FROM ` + tableName +
		` WHERE autoload = ? `

	var optionWithQuote = make([]string, 0)
	for _, option := range optionName {
		optionWithQuote = append(optionWithQuote, fmt.Sprintf("'%s'", option))
	}

	nameString := strings.Join(optionWithQuote, ",")

	if nameString != "" {
		sqlQuery += fmt.Sprintf("OR option_name IN (%s)", nameString)
	}

	q, err := repo.db.Query(sqlQuery, autoload)
	if err != nil {
		log.WithFields(log.Fields{
			"params": fmt.Sprintf("%s, %s, %v", sqlQuery, autoload, optionName),
			"func":   "db.Query",
		}).Errorf("Failed to run db query: %s", err)
		return nil, err
	}
	defer q.Close()

	var res = make([]*model.Option, 0)
	for q.Next() {
		wo := model.Option{}

		err = q.Scan(&wo.OptionID, &wo.OptionName, &wo.OptionValue, &wo.AutoLoad)
		if err != nil {
			log.WithFields(log.Fields{
				"params": wo,
				"func":   "q.Scan",
			}).Errorf("Failed to run db scan: %s", err)
			return nil, err
		}

		res = append(res, &wo)
	}

	return res, nil
}

// LoadOptions is function get an Option from specific option name
func (repo *repository) LoadOption(ctx context.Context, optionName string) (*model.Option, error) {
	apiConfig := ctx.Value(model.APIConfigKey).(model.APIConfig)
	tableName := apiConfig.TablePrefix + "options"

	var sqlQuery = `SELECT ` +
		`option_id, option_name, option_value, autoload ` +
		`FROM ` + tableName +
		` WHERE option_name = ? `

	wo := &model.Option{}
	row := repo.db.QueryRow(sqlQuery, optionName)

	err := row.Scan(&wo.OptionID, &wo.OptionName, &wo.OptionValue, &wo.AutoLoad)

	if err != nil {
		log.WithFields(log.Fields{
			"params": optionName,
			"func":   "row.Scan",
		}).Errorf("Failed to row scan in load option: %s", err)
		return nil, err
	}

	return wo, nil
}

// PostMetasByPostIDs is function get Post Meta from list of integer post id
func (repo *repository) PostMetasByPostIDs(ctx context.Context, idList []uint64) (map[uint64]map[string]string, error) {
	var err error
	apiConfig := ctx.Value(model.APIConfigKey).(model.APIConfig)
	tableName := apiConfig.TablePrefix + "postmeta"
	idParameters := toolbox.UInt64SliceToCSV(idList)

	sqlQuery := fmt.Sprintf(`SELECT meta_id, post_id, meta_key, meta_value FROM %s WHERE post_id IN (%s)`, tableName, idParameters)

	q, err := repo.db.Query(sqlQuery)
	if err != nil {
		log.WithFields(log.Fields{
			"params": sqlQuery,
			"func":   "repo.db.Query(",
		}).Errorf("Failed to run db query: %s", err)
		return nil, err
	}
	defer q.Close()

	// load results
	var res = map[uint64]map[string]string{}
	for q.Next() {
		wp := model.PostMeta{}

		err = q.Scan(&wp.MetaID, &wp.PostID, &wp.MetaKey, &wp.MetaValue)
		if err != nil {
			log.WithFields(log.Fields{
				"params": wp,
				"func":   "row.Scan",
			}).Errorf("Failed to row scan: %s", err)
			return nil, err
		}
		if _, ok := res[wp.PostID]; !ok {
			res[wp.PostID] = map[string]string{wp.MetaKey.String: wp.MetaValue.String}
		} else {
			res[wp.PostID][wp.MetaKey.String] = wp.MetaValue.String
		}

	}

	return res, nil
}
