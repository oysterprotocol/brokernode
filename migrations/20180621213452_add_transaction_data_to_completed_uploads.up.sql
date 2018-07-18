call AddColumnUnlessExists(Database(), 'completed_uploads', 'prl_tx_hash', 'VARCHAR (255) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'completed_uploads', 'prl_tx_nonce', 'bigint (20) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'completed_uploads', 'gas_tx_hash', 'VARCHAR (255) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'completed_uploads', 'gas_tx_nonce', 'bigint (20) DEFAULT NULL');