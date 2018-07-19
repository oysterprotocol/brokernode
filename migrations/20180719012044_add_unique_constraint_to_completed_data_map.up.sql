call DropKeyIfExists(
    Database(),
    'completed_data_maps',
    'completed_data_maps_genesis_hash_chunk_idx_idx',
    1);  # non-unique, 1 for true

call AddConstraintUnlessExists(
    Database(),
    'completed_data_maps',
    'completed_data_maps_genesis_hash_chunk_idx_idx',
    'UNIQUE',
    'UNIQUE KEY (`genesis_hash`, `chunk_idx`)');