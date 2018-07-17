CREATE TABLE IF NOT EXISTS `stored_genesis_hashes` (
  `id` char(36) NOT NULL,
  `genesis_hash` varchar(255) NOT NULL,
  `file_size_bytes` int(11) NOT NULL,
  `num_chunks` int(11) NOT NULL,
  `webnode_count` int(11) DEFAULT NULL,
  `status` int(11) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `stored_genesis_hashes_genesis_hash_idx` (`genesis_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

