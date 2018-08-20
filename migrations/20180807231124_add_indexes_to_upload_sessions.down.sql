call DropColumnIfExists(Database(), 'upload_sessions', 'next_idx_to_attach');
call DropColumnIfExists(Database(), 'upload_sessions', 'next_idx_to_verify');
call DropColumnIfExists(Database(), 'upload_sessions', 'all_data_ready');