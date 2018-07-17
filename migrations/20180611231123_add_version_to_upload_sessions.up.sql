ALTER TABLE `upload_sessions`
  ADD COLUMN IF NOT EXISTS `version` int (10) unsigned DEFAULT 1;