ALTER TABLE `completed_uploads`
  DROP COLUMN IF EXISTS `prl_tx_hash`;

ALTER TABLE `completed_uploads`
  DROP COLUMN IF EXISTS `prl_tx_nonce`;

ALTER TABLE `completed_uploads`
  DROP COLUMN IF EXISTS `gas_tx_hash`;

ALTER TABLE `completed_uploads`
  DROP COLUMN IF EXISTS `gas_tx_nonce`;