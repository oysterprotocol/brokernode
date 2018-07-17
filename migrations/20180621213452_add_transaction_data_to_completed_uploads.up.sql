ALTER TABLE `completed_uploads`
  ADD COLUMN IF NOT EXISTS `prl_tx_hash` VARCHAR (255) DEFAULT NULL;

ALTER TABLE `completed_uploads`
  ADD COLUMN IF NOT EXISTS `prl_tx_nonce` bigint (20) DEFAULT NULL;

ALTER TABLE `completed_uploads`
  ADD COLUMN IF NOT EXISTS `gas_tx_hash` VARCHAR (255) DEFAULT NULL;

ALTER TABLE `completed_uploads`
  ADD COLUMN IF NOT EXISTS `gas_tx_nonce` bigint (20) DEFAULT NULL;