CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    group_name VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    is_income BOOLEAN NOT NULL DEFAULT false,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);
CREATE INDEX idx_categories_user_id ON categories(user_id);
