call DropColumnIfExists(Database(), 'treasures', 'genesis_hash');
call DropColumnIfExists(Database(), 'treasures', 'prl_tx_hash');
call DropColumnIfExists(Database(), 'treasures', 'prl_tx_nonce');
call DropColumnIfExists(Database(), 'treasures', 'gas_tx_hash');
call DropColumnIfExists(Database(), 'treasures', 'gas_tx_nonce');
call DropColumnIfExists(Database(), 'treasures', 'bury_tx_hash');
call DropColumnIfExists(Database(), 'treasures', 'bury_tx_nonce');