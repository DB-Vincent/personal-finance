CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id),
    type VARCHAR(20) NOT NULL CHECK (type IN ('income','expense','transfer')),
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    category_id UUID REFERENCES categories(id),
    transfer_account_id UUID REFERENCES accounts(id),
    date DATE NOT NULL,
    notes TEXT,
    recurring_rule_id UUID,
    created_by UUID NOT NULL,
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID NOT NULL,
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_category_id ON transactions(category_id);
