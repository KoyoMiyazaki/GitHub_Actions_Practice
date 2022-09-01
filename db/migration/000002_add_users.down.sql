ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTs "owner_currency_key";

ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTs "accounts_owner_fkey";

DROP TABLE IF EXISTS "users";