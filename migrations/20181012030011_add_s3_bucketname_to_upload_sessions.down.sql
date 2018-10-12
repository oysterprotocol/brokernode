call DropColumnIfExists(Database(), 'upload_sessions', 'api_version');
call DropColumnIfExists(Database(), 'upload_sessions', 's3_bucket_name');