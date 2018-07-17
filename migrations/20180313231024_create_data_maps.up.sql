CREATE TABLE IF NOT EXISTS `data_maps` (
  `id` char(36) NOT NULL,
  PRIMARY KEY (`id`),
  `genesis_hash` VARCHAR (255) NOT NULL,
  `chunk_idx` integer NOT NULL,
  `hash` VARCHAR (255) NOT NULL,
  `obfuscated_hash` VARCHAR (255) NOT NULL,
  `status` integer NOT NULL,
  `node_id` VARCHAR (255),
  `node_type` VARCHAR (255),
  `message` MEDIUMTEXT,
  `trunk_tx` VARCHAR (255),
  `branch_tx` VARCHAR (255),
  `address` VARCHAR (255),
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  UNIQUE KEY `data_maps_genesis_hash_chunk_idx_idx` (`genesis_hash`, `chunk_idx`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
