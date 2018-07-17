ALTER TABLE `data_maps`
  DROP COLUMN IF EXISTS `msg_id`;

ALTER TABLE `completed_data_maps`
  DROP COLUMN IF EXISTS `msg_id`;

ALTER TABLE `treasures`
  DROP COLUMN IF EXISTS `msg_id`;
