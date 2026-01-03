CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type TEXT NOT NULL, -- e.g., 'income', 'expense'
    amount NUMERIC(15, 2) NOT NULL, -- 15 total digits, 2 after the decimal
    transaction_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign Key Constraint
    CONSTRAINT fk_user 
        FOREIGN KEY(user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

-- Index for faster lookups by UserId
CREATE INDEX idx_transactions_user_id ON transactions(user_id);