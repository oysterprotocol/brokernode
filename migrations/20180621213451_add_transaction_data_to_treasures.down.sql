ALTER TABLE `treasures`
  DROP COLUMN IF EXISTS `genesis_hash`;
ALTER TABLE `treasures`
  DROP COLUMN IF EXISTS `prl_tx_hash`;
ALTER TABLE `treasures`
  DROP COLUMN IF EXISTS `prl_tx_nonce`;
ALTER TABLE `treasures`
  DROP COLUMN IF EXISTS `gas_tx_hash`;
ALTER TABLE `treasures`
  DROP COLUMN IF EXISTS `gas_tx_nonce`;
ALTER TABLE `treasures`
  DROP COLUMN IF EXISTS `bury_tx_hash`;
ALTER TABLE `treasures`
  DROP COLUMN IF EXISTS `bury_tx_nonce`;