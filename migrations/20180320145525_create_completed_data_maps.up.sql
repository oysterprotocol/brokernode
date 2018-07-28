CREATE TABLE IF NOT EXISTS `completed_data_maps` (
  `id`              char(36)     NOT NULL,
  `genesis_hash`    varchar(255) NOT NULL,
  `chunk_idx`       int(11)      NOT NULL,
  `hash`            varchar(255) NOT NULL,
  `obfuscated_hash` varchar(255) NOT NULL,
  `status`          int(11)      NOT NULL,
  `node_id`         varchar(255)          DEFAULT NULL,
  `node_type`       varchar(255)          DEFAULT NULL,
  `message`         text                  DEFAULT NULL,
  `trunk_tx`        varchar(255)          DEFAULT NULL,
  `branch_tx`       varchar(255)          DEFAULT NULL,
  `address`         varchar(255)          DEFAULT NULL,
  `created_at`      datetime     NOT NULL,
  `updated_at`      datetime     NOT NULL,
  PRIMARY KEY (`id`),
  KEY `completed_data_maps_genesis_hash_chunk_idx_idx` (`genesis_hash`, `chunk_idx`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = latin1;
