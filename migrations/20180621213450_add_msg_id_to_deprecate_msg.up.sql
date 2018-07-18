call AddColumnUnlessExists(Database(), 'treasures', 'msg_id', 'VARCHAR (255) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'completed_data_maps', 'msg_id', 'VARCHAR (255) DEFAULT NULL');
call AddColumnUnlessExists(Database(), 'data_maps', 'msg_id', 'VARCHAR (255) NOT NULL');