ALTER TABLE `token`
ADD COLUMN `redeemed_at` DATETIME NULL
AFTER `issued_at`;
