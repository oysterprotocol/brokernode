ALTER TABLE `data_maps`
  ADD COLUMN IF NOT EXISTS `msg_status` int (10) DEFAULT 0;
