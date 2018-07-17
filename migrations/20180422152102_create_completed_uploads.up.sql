CREATE TABLE IF NOT EXISTS `completed_uploads` (
  `id` char(36) NOT NULL,
  `genesis_hash` varchar(255) NOT NULL,
  `eth_addr` varchar(255) NOT NULL,
  `eth_private_key` varchar(255) NOT NULL,
  `prl_status` int(11) NOT NULL,
  `gas_status` int(11) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `version` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `completed_uploads_genesis_hash_idx` (`genesis_hash`),
  UNIQUE KEY `completed_uploads_eth_addr_idx` (`eth_addr`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

