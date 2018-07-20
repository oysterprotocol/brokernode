call DropConstraintIfExists(
    Database(),
    'stored_genesis_hashes',
    'stored_genesis_hashes_genesis_hash_idx',
    'UNIQUE');

call AddKeyUnlessExists(
    Database(),
    'stored_genesis_hashes',
    'stored_genesis_hashes_genesis_hash_idx',
    1, # non-unique, 1 for true
    '(`genesis_hash`)');

call DropConstraintIfExists(
    Database(),
    'brokernodes',
    'brokernodes_address_idx',
    'UNIQUE');