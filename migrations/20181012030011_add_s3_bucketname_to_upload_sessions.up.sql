call AddColumnUnlessExists(Database(), 'upload_sessions', 'storage_method', 'int (10) DEFAULT 2');
call AddColumnUnlessExists(Database(), 'upload_sessions', 's3_bucket_name', 'varchar(63) DEFAULT NULL');
