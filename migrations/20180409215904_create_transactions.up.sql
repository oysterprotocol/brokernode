CREATE TABLE IF NOT EXISTS `transactions` (
	`id` char(36) NOT NULL,
	`status` int(11) NOT NULL,
	`type` int(11) NOT NULL,
	`data_map_id` char(36) NOT NULL,
	`purchase` varchar(255) NOT NULL,
	`created_at` datetime NOT NULL,
	`updated_at` datetime NOT NULL,
	PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;