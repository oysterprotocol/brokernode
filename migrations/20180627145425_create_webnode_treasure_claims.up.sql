CREATE TABLE IF NOT EXISTS `webnode_treasure_claims` (
	`id` char(36) NOT NULL,
	`genesis_hash` varchar(255) NOT NULL,
	`sector_idx` integer NOT NULL,
	`num_chunks` integer NOT NULL,
	`receiver_eth_addr` varchar(255) NOT NULL,
	`treasure_eth_addr` varchar(255) NOT NULL,
	`treasure_eth_private_key` varchar(255) NOT NULL,
	`starting_claim_clock` decimal(28,18) NOT NULL,
	`claim_prl_status` integer NOT NULL,
	`claim_prl_tx_hash` varchar(255) DEFAULT NULL,
	`claim_prl_tx_nonce` decimal(28,18) NOT NULL,
	`gas_status` integer NOT NULL,
	`gas_tx_hash` varchar(255) DEFAULT NULL,
	`gas_tx_nonce` decimal(28,18) NOT NULL,
	PRIMARY KEY (`id`),
	UNIQUE KEY `webnode_treasure_claims_treasure_eth_addr_idx` (`treasure_eth_addr`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;