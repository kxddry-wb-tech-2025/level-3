CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    quantity INTEGER NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    deleted_by UUID REFERENCES users(id),
);

CREATE TABLE IF NOT EXISTS items_history (
    history_id BIGSERIAL PRIMARY KEY,
    action TEXT NOT NULL,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    changed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION log_items_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO items_history (action, item_id, user_id)
        VALUES ('INSERT', NEW.id, NULL);
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
            INSERT INTO items_history (action, item_id, user_id)
            VALUES ('DELETE', NEW.id, NEW.deleted_by);
        ELSE
            INSERT INTO items_history (action, item_id, user_id)
            VALUES ('UPDATE', NEW.id, NULL);
        END IF;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS items_changes_trigger ON items;

CREATE TRIGGER items_changes_trigger
AFTER INSERT OR UPDATE ON items
FOR EACH ROW EXECUTE FUNCTION log_items_changes();

