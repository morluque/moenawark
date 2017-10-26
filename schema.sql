
PRAGMA foreign_keys = ON;

CREATE TABLE characters (
	id INTEGER PRIMARY KEY NOT NULL,
	name TEXT NOT NULL,
	power INTEGER NOT NULL DEFAULT 0,
	actions INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE users (
	id INTEGER PRIMARY KEY NOT NULL,
	email TEXT NOT NULL,
	password TEXT NOT NULL,
	registered BOOLEAN NOT NULL DEFAULT false,
	game_master BOOLEAN NOT NULL DEFAULT false,
	character_id INTEGER DEFAAULT NULL FOREIGN KEY REFERENCES characters(id)
);

CREATE TABLE registrations (
	user_id INTEGER NOT NULL FOREIGN KEY REFERENCES users(id),
	valid_until INTEGER NOT NULL,
	token TEXT NOT NULL,
	status INTEGER NOT NULL DEFAULT 0, -- generated=0, sent=1, error=2
	CONSTRAINT pk_registrations PRIMARY KEY (user_id)
);


CREATE TABLE places (
	id INTEGER PRIMARY KEY NOT NULL,
	name TEXT NOT NULL,
	energy_production INTEGER NOT NULL
);

CREATE TABLE wormholes (
	source_id INTEGER NOT NULL FOREIGN KEY REFERENCES places(id)
	destination_id INTEGER NOT NULL FOREIGN KEY REFERENCES places(id)
	distance INTEGER NOT NULL,
	CONSTRAINT pk_wormholes PRIMARY KEY (source_id, destination_id)
);

CREATE TABLE objects (
	id INTEGER PRIMARY KEY NOT NULL,
	place_id INTEGER NOT NULL FOREIGN KEY REFERENCES places(id)
);

CREATE TABLE atoms (
	id INTEGER PRIMARY KEY NOT NULL,
	name TEXT UNIQUE NOT NULL
);

CREATE TABLE matters (
	id INTEGER PRIMARY KEY NOT NULL,
	quantity INTEGER NOT NULL CONSTRAINT notvoid CHECK (quantity > 0),
	atom_id INTEGER FOREIGN KEY REFERENCES atoms(id),
);

CREATE TABLE resources (
	object_id INTEGER UNIQUE NOT NULL FOREIGN KEY REFERENCES objects(id),
	volume INTEGER NOT NULL CONSTRAINT volume_positive CHECK (quantity > 0),
	sturdiness INTEGER NOT NULL CONSTRAINT sturdiness_positive CHECK (quantity > 0),
	CONSTRAINT pk_resources PRIMARY KEY (object_id)
);

CREATE TABLE resource_components (
	resource_id INTEGER NOT NULL FOREIGN KEY REFERENCES resources(id),
	matter_id INTEGER NOT NULL FOREIGN KEY REFERENCES matters(id),
	CONSTRAINT pk_res_comp PRIMARY KEY (resource_id, matter_id)
);

CREATE TABLE constructions (
	resource_id INTEGER UNIQUE NOT NULL FOREIGN KEY REFERENCES resources(id),
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
	construction_id INTEGER NOT NULL FOREIGN KEY REFERENCES constructions(id),
	object_id INTEGER NOT NULL FOREIGN KEY REFERENCES objects(id),
	CONSTRAINT pk_constr_freight PRIMARY KEY (construction_id, object_id)
);

CREATE TABLE construction_biocompatibility (
	construction_id INTEGER NOT NULL FOREIGN KEY REFERENCES constructions(id),
	matter_id INTEGER NOT NULL FOREIGN KEY REFERENCES matters(id),
	CONSTRAINT pk_constr_bioc PRIMARY KEY (construction_id, matter_id)
);

CREATE TABLE entities (
	resource_id INTEGER NOT NULL FOREIGN KEY REFERENCES resources(id),
	character_id INTEGER DEFAULT NULL FOREIGN KEY REFERENCES characters(id),
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
	target_domain_id INTEGER NOT NULL FOREIGN KEY REFERENCES knowledge_domains(id),
	prerequisite_domain_id INTEGER NOT NULL FOREIGN KEY REFERENCES knowledge_domains(id),
	minimum_proficiency INTEGER NOT NULL,
	CONSTRAINT pk_kn_cstr PRIMARY KEY (target_domain_id, prerequisite_domain_id)
);

CREATE TABLE knowledges (
	entity_id INTEGER NOT NULL FOREIGN KEY REFERENCES entities(id),
	knowledge_domain_id INTEGER NOT NULL FOREIGN KEY REFERENCES knowledge_domains(id),
	proficiency INTEGER NOT NULL
);


CREATE TABLE turns (
	id INTEGER PRIMARY KEY NOT NULL,
	started_at INTEGER NOT NULL,
	ended_at INTEGER DEFAULT NULL
);

CREATE TABLE orders (
	id INTEGER PRIMARY KEY NOT NULL,
	turn_id INTEGER NOT NULL FOREIGN KEY REFERENCES turns(id),
	character_id INTEGER NOT NULL FOREIGN KEY REFERENCES characters(id),
	cost INTEGER NOT NULL,
	order_type TEXT NOT NULL,
	json_args TEXT NOT NULL
);
