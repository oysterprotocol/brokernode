call AddColumnUnlessExists(Database(), 'treasures', 'signed_status', 'int(11) DEFAULT 0');
call AddColumnUnlessExists(Database(), 'treasures', 'encryption_index', 'bigint(20) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'treasures', 'idx', 'bigint(20) DEFAULT NULL');
