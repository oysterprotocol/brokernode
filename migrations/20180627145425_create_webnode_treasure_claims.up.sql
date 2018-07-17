CREATE TABLE IF NOT EXISTS `webnode_treasure_claims` (
  `id`                       char(36)     NOT NULL,
  `created_at`               datetime     NOT NULL,
  `updated_at`               datetime     NOT NULL,
  `genesis_hash`             varchar(255) NOT NULL,
  `sector_idx`               int(11)      NOT NULL,
  `num_chunks`               int(11)      NOT NULL,
  `receiver_eth_addr`        varchar(255) NOT NULL,
  `treasure_eth_addr`        varchar(255) NOT NULL,
  `treasure_eth_private_key` varchar(255) NOT NULL,
  `starting_claim_clock`     bigint(20)   NOT NULL,
  `claim_prl_status`         int(11)      NOT NULL,
  `claim_prl_tx_hash`        varchar(255) DEFAULT NULL,
  `claim_prl_tx_nonce`       bigint(20)   NOT NULL,
  `gas_status`               int(11)      NOT NULL,
  `gas_tx_hash`              varchar(255) DEFAULT NULL,
  `gas_tx_nonce`             bigint(20)   NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `webnode_treasure_claims_treasure_eth_addr_idx` (`treasure_eth_addr`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = latin1;