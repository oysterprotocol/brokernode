ALTER TABLE `stored_genesis_hashes`
  ADD COLUMN IF NOT EXISTS `treasure_status` int (10) DEFAULT 1;