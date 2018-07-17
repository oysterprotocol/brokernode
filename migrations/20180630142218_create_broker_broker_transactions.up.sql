CREATE TABLE IF NOT EXISTS `broker_broker_transactions` (
  `id`              char(36)        NOT NULL,
  `created_at`      datetime        NOT NULL,
  `updated_at`      datetime        NOT NULL,
  `genesis_hash`    varchar(255)    NOT NULL,
  `type`            int(11)         NOT NULL,
  `eth_addr_alpha`  varchar(255)    NOT NULL,
  `eth_addr_beta`   varchar(255)    NOT NULL,
  `eth_private_key` varchar(255) DEFAULT NULL,
  `total_cost`      decimal(28, 18) NOT NULL,
  `payment_status`  int(11)         NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `broker_broker_transactions_genesis_hash_idx` (`genesis_hash`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = latin1;