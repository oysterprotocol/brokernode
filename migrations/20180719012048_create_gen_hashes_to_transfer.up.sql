CREATE TABLE IF NOT EXISTS `transfer_gen_hashes` (
  `id`           char(36)     NOT NULL,
  `genesis_hash` varchar(255) NOT NULL,
  `num_chunks`   int(11)      NOT NULL,
  `complete`     int(11) DEFAULT 0,
  `created_at`   datetime     NOT NULL,
  `updated_at`   datetime     NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `transfer_gen_hashes_genesis_hash_idx` (`genesis_hash`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = latin1;