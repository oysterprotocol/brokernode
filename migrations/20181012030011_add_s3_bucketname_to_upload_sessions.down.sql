call DropColumnIfExists(Database(), 'upload_sessions', 'storage_method');
call DropColumnIfExists(Database(), 'upload_sessions', 's3_bucket_name');
