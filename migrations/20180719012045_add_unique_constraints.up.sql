call DropKeyIfExists(
    Database(),
    'stored_genesis_hashes',
    'stored_genesis_hashes_genesis_hash_idx',
    1);  # non-unique, 1 for true

call AddConstraintUnlessExists(
    Database(),
    'stored_genesis_hashes',
    'stored_genesis_hashes_genesis_hash_idx',
    'UNIQUE',
    'UNIQUE KEY (`genesis_hash`)');

call AddConstraintUnlessExists(
    Database(),
    'brokernodes',
    'brokernodes_address_idx',
    'UNIQUE',
    'UNIQUE KEY (`address`)');