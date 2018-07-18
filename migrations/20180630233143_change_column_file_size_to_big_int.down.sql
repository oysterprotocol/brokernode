call ChangeColumnIfExists(Database(), 'stored_genesis_hashes', 'file_size_bytes', 'int(11) NOT NULL');
call ChangeColumnIfExists(Database(), 'upload_sessions', 'file_size_bytes', 'int(11) NOT NULL');