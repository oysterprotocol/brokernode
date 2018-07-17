CREATE TABLE IF NOT EXISTS `webnodes` (
  `id`         char(36)     NOT NULL,
  `address`    varchar(255) NOT NULL,
  `created_at` datetime     NOT NULL,
  `updated_at` datetime     NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `webnodes_address_idx` (`address`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = latin1;

