ALTER TABLE `completed_uploads`
  ADD COLUMN IF NOT EXISTS `prl_tx_hash` VARCHAR (255) DEFAULT NULL;

ALTER TABLE `completed_uploads`
  ADD COLUMN IF NOT EXISTS `prl_tx_nonce` decimal (28, 18) DEFAULT NULL;

ALTER TABLE `completed_uploads`
  ADD COLUMN IF NOT EXISTS `gas_tx_hash` VARCHAR (255) DEFAULT NULL;

ALTER TABLE `completed_uploads`
  ADD COLUMN IF NOT EXISTS `gas_tx_nonce` decimal (28, 18) DEFAULT NULL;