call DropConstraintIfExists(
    Database(),
    'completed_data_maps',
    'completed_data_maps_genesis_hash_chunk_idx_idx',
    'UNIQUE');

call AddKeyUnlessExists(
    Database(),
    'completed_data_maps',
    'completed_data_maps_genesis_hash_chunk_idx_idx',
    1, # non-unique, 1 for true
    '(`genesis_hash`, `chunk_idx`)');