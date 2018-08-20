call AddColumnUnlessExists(Database(), 'upload_sessions', 'next_idx_to_attach', 'bigint(20) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'upload_sessions', 'next_idx_to_verify', 'bigint(20) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'upload_sessions', 'all_data_ready', 'int(11) DEFAULT 1');