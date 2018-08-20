call AddColumnUnlessExists(Database(), 'transactions', 'idx', 'bigint(20) NOT NULL');
call AddColumnUnlessExists(Database(), 'transactions', 'genesis_hash', 'VARCHAR (255) NOT NULL');