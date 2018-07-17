ALTER TABLE `stored_genesis_hashes`
  MODIFY COLUMN IF EXISTS `file_size_bytes` BIGINT ZEROFILL NOT NULL;

ALTER TABLE `upload_sessions`
  MODIFY COLUMN IF EXISTS `file_size_bytes` BIGINT ZEROFILL NOT NULL;
