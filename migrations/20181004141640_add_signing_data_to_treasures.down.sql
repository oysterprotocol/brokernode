call DropColumnIfExists(Database(), 'treasures', 'signed_status');
call DropColumnIfExists(Database(), 'treasures', 'encryption_index');
call DropColumnIfExists(Database(), 'treasures', 'idx');