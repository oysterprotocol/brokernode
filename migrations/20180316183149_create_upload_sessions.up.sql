CREATE TABLE IF NOT EXISTS `upload_sessions` (
	`id` char(36) NOT NULL,
	`genesis_hash` varchar(255) NOT NULL,
	`file_size_bytes` int(11) NOT NULL,
	`num_chunks` int(11) NOT NULL,
	`storage_length_in_years` int(11) NOT NULL,
	`type` int(11) NOT NULL,
	`eth_addr_alpha` varchar(255) DEFAULT NULL,
	`eth_addr_beta` varchar(255) DEFAULT NULL,
	`eth_private_key` varchar(255) DEFAULT NULL,
	`total_cost` decimal(28,18) NOT NULL,
	`payment_status` int(11) NOT NULL,
	`treasure_status` int(11) NOT NULL,
	`treasure_idx_map` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
	`created_at` datetime NOT NULL,
	`updated_at` datetime NOT NULL,
	`version` int(10) unsigned DEFAULT NULL,
	PRIMARY KEY (`id`),
	UNIQUE KEY `upload_sessions_genesis_hash_idx` (`genesis_hash`),
	UNIQUE KEY `upload_sessions_eth_addr_alpha_idx` (`eth_addr_alpha`),
	UNIQUE KEY `upload_sessions_eth_addr_beta_idx` (`eth_addr_beta`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

