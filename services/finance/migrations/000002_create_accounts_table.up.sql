CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('checking','savings','credit_card','cash','investment','loan','other')),
    starting_balance NUMERIC(12,2) NOT NULL DEFAULT 0,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    created_by UUID NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID NOT NULL,
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_accounts_user_id ON accounts(user_id);
