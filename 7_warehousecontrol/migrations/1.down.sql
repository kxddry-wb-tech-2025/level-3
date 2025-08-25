
DROP TRIGGER IF EXISTS items_changes_trigger ON items;
DROP FUNCTION IF EXISTS log_items_changes();
DROP TABLE IF EXISTS items_history;

DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";