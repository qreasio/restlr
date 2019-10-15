package term

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/qreasio/restlr/model"
	"github.com/qreasio/restlr/toolbox"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Repository interface {
	PostTermTaxonomyByIDs(ctx context.Context, idStringArray []string) (map[uint64][]*model.TermWithPostTaxonomy, error)
	GetPostTaxonomyAndFormat(ctx context.Context, idStringArr []string) (map[uint64][]*model.TermWithPostTaxonomy, map[uint64]map[string][]uint64, map[uint64]string, error)
	TermTaxonomyByTermIDListTaxonomy(termIDList []uint64, taxonomy string) ([]*model.TermTaxonomy, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (repo *repository) PostTermTaxonomyByIDs(ctx context.Context, idStringArray []string) (map[uint64][]*model.TermWithPostTaxonomy, error) {
	config := ctx.Value(model.APICONFIGKEY).(model.APIModel)
	termsTableName := config.TablePrefix + "terms"
	termTaxonomyTableName := config.TablePrefix + "term_taxonomy"
	termRelationsTableName := config.TablePrefix + "term_relationships"
	sql := fmt.Sprintf(`SELECT t.*, tt.term_taxonomy_id, tt.taxonomy, tt.description, tt.parent, tt.count, tr.object_id FROM %s AS t  `+
		`INNER JOIN %s AS tt ON t.term_id = tt.term_id `+
		`INNER JOIN %s AS tr ON tr.term_taxonomy_id = tt.term_taxonomy_id `+
		`WHERE tt.taxonomy IN ('category', 'post_tag', 'post_format') AND tr.object_id IN (%s) `+
		`ORDER BY t.name ASC`,
		termsTableName,
		termTaxonomyTableName,
		termRelationsTableName,
		strings.Join(idStringArray, ","))

	q, err := repo.db.Query(sql)

	if err != nil {
		log.WithFields(log.Fields{
			"params": sql,
			"func":   "db.Query",
		}).Errorf("Failed to run db query: %s", err)
		return nil, err
	}
	defer q.Close()

	// load results
	result := map[uint64][]*model.TermWithPostTaxonomy{}

	for q.Next() {
		t := model.TermWithPostTaxonomy{}
		// scan
		err = q.Scan(&t.TermID, &t.Name, &t.Slug, &t.TermGroup, &t.TermTaxonomyID, &t.Taxonomy, &t.Description, &t.Parent, &t.Count, &t.ObjectID)

		if err != nil {
			log.WithFields(log.Fields{
				"params": sql,
				"func":   "q.Scan",
			}).Errorf("Failed to run q scan: %s", err)
			return nil, err
		}

		_, exists := result[t.ObjectID]
		if !exists {
			result[t.ObjectID] = []*model.TermWithPostTaxonomy{}
		}

		result[t.ObjectID] = append(result[t.ObjectID], &t)
	}
	return result, nil
}

func (repo *repository) GetPostTaxonomyAndFormat(ctx context.Context, idStringArr []string) (map[uint64][]*model.TermWithPostTaxonomy, map[uint64]map[string][]uint64, map[uint64]string, error) {
	termsMap, err := repo.PostTermTaxonomyByIDs(ctx, idStringArr)

	if err != nil {
		log.WithFields(log.Fields{
			"params": idStringArr,
			"func":   "repo.PostTermTaxonomyByIDs",
		}).Errorf("Failed to get PostTermTaxonomyByIDs: %s", err)
		return nil, nil, nil, err
	}

	taxonomies := map[uint64]map[string][]uint64{}
	formatMap := map[uint64]string{}

	for _, terms := range termsMap {
		for _, t := range terms {
			if t.Taxonomy == "post_tag" || t.Taxonomy == "category" {
				_, exists := taxonomies[t.ObjectID]
				if !exists {
					taxonomies[t.ObjectID] = make(map[string][]uint64)
				}
				_, ok := taxonomies[t.ObjectID][t.Taxonomy]
				if !ok {
					taxonomies[t.ObjectID][t.Taxonomy] = []uint64{}
				}
				taxonomies[t.ObjectID][t.Taxonomy] = append(taxonomies[t.ObjectID][t.Taxonomy], t.TermID)
			}

			if t.Taxonomy == "post_format" {
				formatMap[t.ObjectID] = t.Name
			}
		}
	}

	return termsMap, taxonomies, formatMap, nil
}

// TermTaxonomyByTermIDListTaxonomy get array of TermTaxonomy from specified term ID slices
func (repo *repository) TermTaxonomyByTermIDListTaxonomy(termIDList []uint64, taxonomy string) ([]*model.TermTaxonomy, error) {
	if len(termIDList) == 0 {
		return []*model.TermTaxonomy{}, nil
	}

	sqlQuery := `SELECT ` +
		`term_taxonomy_id, term_id, taxonomy, description, parent, count ` +
		`FROM wp_term_taxonomy ` +
		`WHERE term_id IN ( ` + toolbox.UInt64SliceToCSV(termIDList) + `)  AND taxonomy = ?`

	q, err := repo.db.Query(sqlQuery, taxonomy)
	if err != nil {
		log.WithFields(log.Fields{
			"params": fmt.Sprintf("%s, %v, %s", sqlQuery, termIDList, taxonomy),
			"func":   "db.Query",
		}).Errorf("Failed to run db query: %s", err)
		return nil, err
	}
	defer q.Close()

	// load results
	var res = make([]*model.TermTaxonomy, 0)
	for q.Next() {
		var wtt model.TermTaxonomy

		// scan
		err = q.Scan(&wtt.TermTaxonomyID, &wtt.TermID, &wtt.Taxonomy, &wtt.Description, &wtt.Parent, &wtt.Count)
		if err != nil {
			log.WithFields(log.Fields{
				"params": termIDList,
				"func":   "q.Scan",
			}).Errorf("Failed to run query scan: %s", err)
			return nil, err
		}

		res = append(res, &wtt)
	}

	return res, nil
}
