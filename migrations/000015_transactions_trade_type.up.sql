-- Trade settlement (SettleTrade) records ledger movements with type 'trade', but
-- the original transactions.type CHECK constraint (000001) omitted it, so every
-- trade execution failed with a check-constraint violation even though the order
-- placed fine. Widen the constraint to include 'trade'.
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_type_check;
ALTER TABLE transactions ADD CONSTRAINT transactions_type_check
    CHECK (type IN ('deposit', 'withdrawal', 'transfer', 'fee', 'reversal', 'trade'));
