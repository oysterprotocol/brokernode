CREATE TABLE IF NOT EXISTS `transfer_addresses` (
  `id`           char(36)     NOT NULL,
  `genesis_hash` varchar(255) NOT NULL,
  `chunk_idx`    int(11)      NOT NULL,
  `hash`         varchar(255) NOT NULL,
  `message`      mediumtext   DEFAULT NULL,
  `address`      varchar(255) DEFAULT NULL,
  `created_at`   datetime     NOT NULL,
  `updated_at`   datetime     NOT NULL,
  `complete`     int(11)      DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `transfer_addresses_address_idx` (`address`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = latin1;