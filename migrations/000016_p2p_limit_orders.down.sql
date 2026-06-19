ALTER TABLE p2p_advertisements ADD COLUMN IF NOT EXISTS min_amount NUMERIC(36,18) DEFAULT 0;
ALTER TABLE p2p_advertisements ADD COLUMN IF NOT EXISTS max_amount NUMERIC(36,18) DEFAULT 0;

UPDATE p2p_advertisements SET min_amount = 0, max_amount = total_amount;

ALTER TABLE p2p_advertisements DROP COLUMN IF EXISTS type;
ALTER TABLE p2p_advertisements DROP COLUMN IF EXISTS total_amount;
ALTER TABLE p2p_advertisements DROP COLUMN IF EXISTS remaining_amount;
