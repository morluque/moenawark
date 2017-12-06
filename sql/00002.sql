CREATE TABLE new_mwk_schema_versions (
	id INTEGER PRIMARY KEY NOT NULL,
	num INTEGER UNIQUE NOT NULL,
	deployed_at INTEGER NOT NULL
);
INSERT INTO new_mwk_schema_versions (id, num, deployed_at)
	SELECT id, num, strftime('%s', '1970-01-01T00:00:00') FROM mwk_schema_versions;

DROP TABLE mwk_schema_versions;
ALTER TABLE new_mwk_schema_versions RENAME TO mwk_schema_versions;

INSERT INTO mwk_schema_versions (num, deployed_at) VALUES (2, strftime('%s', 'now'));

PRAGMA foreign_key_check;
