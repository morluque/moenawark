/*
Package sqlstore is a collection of utility functions for database handling.
*/
package sqlstore

import (
	"database/sql"
	"fmt"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/loglevel"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

var log *loglevel.Logger

func init() {
	log = loglevel.New("sqlstore", loglevel.Debug)
}

// ReloadConfig performs required actions to reload all dynamic config.
func ReloadConfig() {
	log.SetLevelName(config.Get("loglevel.sqlstore"))
}

// Open returns a new database connection, and upgrades the schema as needed.
// Currently uses sqlite.
func Open(dbPath string) (*sql.DB, error) {
	// Get current schema version (0 if database not created)
	db, currentVersion, err := openPath(dbPath)
	if err != nil {
		return nil, err
	}
	log.Debugf("Current schema version is %d", currentVersion)

	// Get ordered list of new versions to deploy.
	sqlToExec, lastVersion, err := getVersionsToDeploy(currentVersion)
	if err != nil {
		return nil, err
	}
	if currentVersion == lastVersion {
		log.Debugf("Database schema version is up-to-date.")
		return db, nil
	}

	// Update database with each new version in turn.
	log.Warnf("Database schema version is %d, needs upgrading to %d", currentVersion, lastVersion)
	err = updateDatabaseSchema(db, sqlToExec)
	if err != nil {
		return nil, err
	}
	log.Warnf("Database schema upgraded to version %d", lastVersion)

	return db, nil
}

// IsConstraintError returns true if err is a constraint violation error.
func IsConstraintError(err error) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(sqlite3.Error); ok {
		switch e.Code {
		case sqlite3.ErrConstraint:
			return true
		default:
			return false
		}
	}
	return false
}

func getSchemaVersion(db *sql.DB) (int, error) {
	// Get current version from version table.
	row := db.QueryRow(`
		SELECT num
		  FROM mwk_schema_versions
		 WHERE id = (SELECT max(id) FROM mwk_schema_versions)`)
	var v int
	err := row.Scan(&v)
	if err != nil {
		log.Debugf("SELECT version from mwk_schema_version: %s", err.Error())
		return 0, err
	}
	return v, nil
}

type sortByMapValues map[int]string

func (m sortByMapValues) Len() int           { return len(m) }
func (m sortByMapValues) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m sortByMapValues) Less(i, j int) bool { return i < j }

func getVersionsToDeploy(currentVersion int) ([]string, int, error) {
	sqlDir := config.Get("sql_path")
	sqlPath, err := os.Open(sqlDir)
	if err != nil {
		return nil, 0, err
	}
	files, err := sqlPath.Readdirnames(-1)
	if err != nil {
		return nil, 0, err
	}

	sqlFiles := make(map[int]string)
	re, err := regexp.Compile("^([0-9]{5})[.]sql$")
	if err != nil {
		panic("WTFBBQ!!!11! bad regexp for matching SQL files.")
	}
	for _, f := range files {
		matches := re.FindStringSubmatch(f)
		if matches == nil {
			continue
		}
		v, err := strconv.Atoi(matches[1])
		if err != nil {
			panic("WTFBBQ!!!11! can't convert sql version str to int")
		}
		sqlFiles[v] = f
		log.Debugf("Found SQL file %s for version %d", f, v)
	}
	log.Debugf("Found %d SQL file(s)", len(sqlFiles))
	sort.Sort(sortByMapValues(sqlFiles))

	sqlToExec := make([]string, 0)
	last := 0
	for v, f := range sqlFiles {
		last = v
		if v <= currentVersion {
			continue
		}
		sqlToExec = append(sqlToExec, filepath.Join(sqlDir, f))
	}

	return sqlToExec, last, nil
}

func updateDatabaseSchema(db *sql.DB, sqlFiles []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, fPath := range sqlFiles {
		log.Debugf("Opening SQL file %s", fPath)
		f, err := os.Open(fPath)
		if err != nil {
			return err
		}
		info, err := f.Stat()
		if err != nil {
			return err
		}
		log.Debugf("Reading SQL file %s", fPath)
		sql := make([]byte, info.Size())
		_, err = io.ReadFull(f, sql)
		if err != nil {
			return err
		}
		log.Infof("Executing SQL in file %s", fPath)
		_, err = tx.Exec(string(sql))
		if err != nil {
			return err
		}
	}
	log.Debugf("Committing upgrade transaction")
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func openPath(dbPath string) (*sql.DB, int, error) {
	dataSourceName := fmt.Sprintf("file:%s?_busy_timeout=1000&_foreign_keys=1", dbPath)
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, 0, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, 0, err
	}
	v, err := getSchemaVersion(db)
	if err != nil {
		return db, 0, nil
	}
	return db, v, nil
}
