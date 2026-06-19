ALTER TABLE p2p_advertisements ADD COLUMN IF NOT EXISTS type VARCHAR(10) DEFAULT 'sell';
ALTER TABLE p2p_advertisements ADD COLUMN IF NOT EXISTS total_amount NUMERIC(36,18) DEFAULT 0;
ALTER TABLE p2p_advertisements ADD COLUMN IF NOT EXISTS remaining_amount NUMERIC(36,18) DEFAULT 0;

-- Migrate existing data
UPDATE p2p_advertisements SET type = 'sell', total_amount = max_amount, remaining_amount = max_amount;

-- We can leave min_amount and max_amount columns or drop them. Let's drop them to keep schema clean.
ALTER TABLE p2p_advertisements DROP COLUMN IF EXISTS min_amount;
ALTER TABLE p2p_advertisements DROP COLUMN IF EXISTS max_amount;
