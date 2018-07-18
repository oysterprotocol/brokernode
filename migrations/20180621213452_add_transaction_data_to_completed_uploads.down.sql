call DropColumnIfExists(Database(), 'completed_uploads', 'prl_tx_hash');
call DropColumnIfExists(Database(), 'completed_uploads', 'prl_tx_nonce');
call DropColumnIfExists(Database(), 'completed_uploads', 'gas_tx_hash');
call DropColumnIfExists(Database(), 'completed_uploads', 'gas_tx_nonce');