ALTER TABLE `standings`
ADD COLUMN `finalised` BOOLEAN NOT NULL
AFTER rankings;
