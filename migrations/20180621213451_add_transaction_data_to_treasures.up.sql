call AddColumnUnlessExists(Database(), 'treasures', 'genesis_hash', 'VARCHAR (255) NOT NULL');
call AddColumnUnlessExists(Database(), 'treasures', 'prl_tx_hash', 'VARCHAR (255) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'treasures', 'prl_tx_nonce', 'bigint (20) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'treasures', 'gas_tx_hash', 'VARCHAR (255) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'treasures', 'gas_tx_nonce', 'bigint (20) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'treasures', 'bury_tx_hash', 'VARCHAR (255) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'treasures', 'bury_tx_nonce', 'bigint (20) DEFAULT NULL');