CREATE TABLE IF NOT EXISTS `treasures` (
  `id`         char(36)     NOT NULL,
  `eth_addr`   varchar(255) NOT NULL,
  `eth_key`    varchar(255) NOT NULL,
  `prl_amount` varchar(255) NOT NULL,
  `prl_status` int(11)      NOT NULL,
  `message`    mediumtext   DEFAULT NULL,
  `address`    varchar(255) DEFAULT NULL,
  `created_at` datetime     NOT NULL,
  `updated_at` datetime     NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `treasures_eth_addr_idx` (`eth_addr`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = latin1;
