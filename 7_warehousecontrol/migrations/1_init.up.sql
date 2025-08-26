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
    created_by UUID REFERENCES users(id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID REFERENCES users(id) ON DELETE CASCADE,
    deleted_at TIMESTAMPTZ,
    deleted_by UUID REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS items_history (
    id BIGSERIAL PRIMARY KEY,
    action TEXT NOT NULL,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    role INTEGER NOT NULL,
    changed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    old_data JSONB,
    new_data JSONB
);


CREATE OR REPLACE FUNCTION log_items_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO items_history (action, item_id, user_id, role, old_data, new_data)
        VALUES (
            'INSERT',
            NEW.id,
            NEW.created_by,
            (SELECT role FROM users WHERE id = NEW.created_by),
            NULL,
            to_jsonb(NEW)
        );
        RETURN NEW;

    ELSIF TG_OP = 'UPDATE' THEN
        IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
            INSERT INTO items_history (action, item_id, user_id, role, old_data, new_data)
            VALUES (
                'DELETE',
                NEW.id,
                NEW.deleted_by,
                (SELECT role FROM users WHERE id = NEW.deleted_by),
                to_jsonb(OLD),
                to_jsonb(NEW)
            );
        ELSE
            INSERT INTO items_history (action, item_id, user_id, role, old_data, new_data)
            VALUES (
                'UPDATE',
                NEW.id,
                NEW.updated_by,
                (SELECT role FROM users WHERE id = NEW.updated_by),
                to_jsonb(OLD),
                to_jsonb(NEW)
            );
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

