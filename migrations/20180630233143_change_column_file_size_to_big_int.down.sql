ALTER TABLE `stored_genesis_hashes`
  MODIFY COLUMN IF EXISTS `file_size_bytes` int(11) NOT NULL;

ALTER TABLE `upload_sessions`
  MODIFY COLUMN IF EXISTS `file_size_bytes` int(11) NOT NULL;