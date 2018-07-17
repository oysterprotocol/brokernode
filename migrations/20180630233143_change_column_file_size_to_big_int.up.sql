ALTER TABLE `stored_genesis_hashes`
  MODIFY COLUMN IF EXISTS `file_size_bytes` bigint(20) unsigned zerofill NOT NULL;

ALTER TABLE `upload_sessions`
  MODIFY COLUMN IF EXISTS `file_size_bytes` bigint(20) unsigned zerofill NOT NULL;
