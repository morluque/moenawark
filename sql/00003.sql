PRAGMA foreign_keys=OFF;

CREATE TABLE new_users (
	id INTEGER PRIMARY KEY NOT NULL,
	created_at INTEGER NOT NULL,
	login TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'new' CHECK (status in ('new', 'active', 'archived')),
	game_master BOOLEAN NOT NULL DEFAULT false,
	character_id INTEGER DEFAULT NULL CONSTRAINT fk_user_char REFERENCES characters(id)
);
INSERT INTO new_users (id, login, password, status, game_master, character_id, created_at)
	SELECT id, login, password, status, game_master, character_id, strftime('%s', '1970-01-01T00:00:00')
	  FROM users;

DROP TABLE users;
ALTER TABLE new_users RENAME TO users;

PRAGMA foreign_key_check;
PRAGMA foreign_keys=ON;

INSERT INTO mwk_schema_versions (num, deployed_at) VALUES (3, strftime('%s', 'now'));
