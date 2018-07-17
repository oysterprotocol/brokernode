CREATE TABLE IF NOT EXISTS `chunk_channels` (
  `id`               char(36)     NOT NULL,
  `channel_id`       varchar(255) NOT NULL,
  `chunks_processed` int(11)      NOT NULL,
  `est_ready_time`   datetime DEFAULT NULL,
  `created_at`       datetime     NOT NULL,
  `updated_at`       datetime     NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `chunk_channels_channel_id_idx` (`channel_id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = latin1;

