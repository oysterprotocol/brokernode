ALTER TABLE `completed_uploads`
  ADD COLUMN IF NOT EXISTS `version` int (10) unsigned DEFAULT 1;