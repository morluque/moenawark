package sqlstore

import (
	"database/sql"
	"fmt"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/morluque/moenawark/mwkerr"
	"os"
)

// DB is a common interface between *sql.Tx and *sql.DB
type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
}

// Open returns a new database connection.
// Currently uses sqlite.
func Open(dbPath string) (*sql.DB, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, mwkerr.New(mwkerr.DatabaseEmpty, "Database %s does not exist, must be initialized", dbPath)
	}

	db, err := openPath(dbPath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Init creates a new empty database with our SQL schema
func Init(dbPath string) (*sql.DB, error) {
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		return nil, mwkerr.New(mwkerr.DatabaseAlreadyInitialized, "Can't initialize existing database %s", dbPath)
	}

	db, err := openPath(dbPath)
	if err != nil {
		return nil, err
	}
	err = create(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func openPath(dbPath string) (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("file:%s", dbPath)
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

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

func create(db *sql.DB) error {
	ddml := `
PRAGMA foreign_keys = ON;

CREATE TABLE characters (
	id INTEGER PRIMARY KEY NOT NULL,
	name TEXT UNIQUE NOT NULL,
	power INTEGER NOT NULL DEFAULT 0,
	actions INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE users (
	id INTEGER PRIMARY KEY NOT NULL,
	login TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'new' CHECK (status in ('new', 'active', 'archived')),
	game_master BOOLEAN NOT NULL DEFAULT false,
	character_id INTEGER DEFAULT NULL CONSTRAINT fk_user_char REFERENCES characters(id)
);
CREATE INDEX user_status_idx ON users (status);


CREATE TABLE places (
	id INTEGER PRIMARY KEY NOT NULL,
	name TEXT NOT NULL UNIQUE,
	x INTEGER NOT NULL,
	y INTEGER NOT NULL,
	energy_production INTEGER NOT NULL
);
CREATE UNIQUE INDEX place_pos_idx ON places (x, y);

CREATE TABLE wormholes (
	source_id INTEGER NOT NULL CONSTRAINT fk_wormh_source REFERENCES places(id),
	destination_id INTEGER NOT NULL CONSTRAINT fk_wormh_dest REFERENCES places(id),
	distance INTEGER NOT NULL,
	CONSTRAINT pk_wormholes PRIMARY KEY (source_id, destination_id)
);

CREATE TABLE objects (
	id INTEGER PRIMARY KEY NOT NULL,
	place_id INTEGER NOT NULL CONSTRAINT fk_obj_place REFERENCES places(id)
);

CREATE TABLE atoms (
	id INTEGER PRIMARY KEY NOT NULL,
	name TEXT UNIQUE NOT NULL
);

CREATE TABLE matters (
	id INTEGER PRIMARY KEY NOT NULL,
	quantity INTEGER NOT NULL CONSTRAINT notvoid CHECK (quantity > 0),
	atom_id INTEGER NOT NULL CONSTRAINT fk_matter_atom REFERENCES atoms(id)
);

CREATE TABLE resources (
	object_id INTEGER UNIQUE NOT NULL CONSTRAINT fk_res_obj REFERENCES objects(id),
	volume INTEGER NOT NULL CONSTRAINT volume_positive CHECK (volume > 0),
	sturdiness INTEGER NOT NULL CONSTRAINT sturdiness_positive CHECK (sturdiness > 0),
	CONSTRAINT pk_resources PRIMARY KEY (object_id)
);

CREATE TABLE resource_components (
	resource_id INTEGER NOT NULL CONSTRAINT fk_res_comp_res REFERENCES resources(id),
	matter_id INTEGER NOT NULL CONSTRAINT fk_res_comp_matter REFERENCES matters(id),
	CONSTRAINT pk_res_comp PRIMARY KEY (resource_id, matter_id)
);

CREATE TABLE constructions (
	resource_id INTEGER UNIQUE NOT NULL CONSTRAINT fk_constr_res REFERENCES resources(id),
	group_count INTEGER NOT NULL DEFAULT 1,
	name TEXT NOT NULL,
	attack INTEGER NOT NULL DEFAULT 0,
	movement INTEGER NOT NULL DEFAULT 0,
	storage_volume INTEGER NOT NULL DEFAULT 0,
	energy_level INTEGER NOT NULL DEFAULT 0,
	energy_storage INTEGER NOT NULL DEFAULT 0,
	energy_harvesting INTEGER NOT NULL DEFAULT 0,
	CONSTRAINT pk_constructions PRIMARY KEY (resource_id)
);

CREATE TABLE construction_freight (
	construction_id INTEGER NOT NULL CONSTRAINT fk_constr_freight_constr REFERENCES constructions(id),
	object_id INTEGER NOT NULL CONSTRAINT fk_constr_freight_obj REFERENCES objects(id),
	CONSTRAINT pk_constr_freight PRIMARY KEY (construction_id, object_id)
);

CREATE TABLE construction_biocompatibility (
	construction_id INTEGER NOT NULL CONSTRAINT fk_constr_bioc_constr REFERENCES constructions(id),
	matter_id INTEGER NOT NULL CONSTRAINT fk_constr_bioc_matter REFERENCES matters(id),
	CONSTRAINT pk_constr_bioc PRIMARY KEY (construction_id, matter_id)
);

CREATE TABLE entities (
	resource_id INTEGER NOT NULL CONSTRAINT fk_ent_res REFERENCES resources(id),
	character_id INTEGER DEFAULT NULL CONSTRAINT fk_ent_char REFERENCES characters(id),
	group_count INTEGER NOT NULL DEFAULT 1,
	name TEXT NOT NULL,
	trust INTEGER NOT NULL DEFAULT 0,
	energy_level INTEGER NOT NULL DEFAULT 0,
	energy_storage INTEGER NOT NULL DEFAULT 0,
	energy_harvesting INTEGER NOT NULL DEFAULT 0,
	CONSTRAINT pk_entities PRIMARY KEY (resource_id)
);

CREATE TABLE knowledge_domains (
	id INTEGER PRIMARY KEY NOT NULL,
	name TEXT UNIQUE NOT NULL
);

CREATE TABLE knowledge_constraints (
	target_domain_id INTEGER NOT NULL CONSTRAINT fk_knwl_constr_target REFERENCES knowledge_domains(id),
	prerequisite_domain_id INTEGER NOT NULL CONSTRAINT fk_knwl_constr_prereq REFERENCES knowledge_domains(id),
	minimum_proficiency INTEGER NOT NULL,
	CONSTRAINT pk_kn_cstr PRIMARY KEY (target_domain_id, prerequisite_domain_id)
);

CREATE TABLE knowledges (
	entity_id INTEGER NOT NULL CONSTRAINT fk_knwl_ent REFERENCES entities(id),
	knowledge_domain_id INTEGER NOT NULL CONSTRAINT fk_knwl_dom REFERENCES knowledge_domains(id),
	proficiency INTEGER NOT NULL
);


CREATE TABLE turns (
	id INTEGER PRIMARY KEY NOT NULL,
	started_at INTEGER NOT NULL,
	ended_at INTEGER DEFAULT NULL
);

CREATE TABLE orders (
	id INTEGER PRIMARY KEY NOT NULL,
	turn_id INTEGER NOT NULL CONSTRAINT fk_order_turn REFERENCES turns(id),
	character_id INTEGER NOT NULL CONSTRAINT fk_order_char REFERENCES characters(id),
	cost INTEGER NOT NULL,
	order_type TEXT NOT NULL,
	json_args TEXT NOT NULL
);
`
	_, err := db.Exec(ddml)
	if err != nil {
		db.Close()
		return err
	}
	return nil
}
