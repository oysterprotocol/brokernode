-- Copyright (c) 2009 www.cryer.co.uk
-- Script is free to use provided this copyright header is included.
drop procedure if exists AddColumnUnlessExists;
# DELIMITER '$$'

create procedure AddColumnUnlessExists(
  IN dbName tinytext,
  IN tableName tinytext,
  IN fieldName tinytext,
  IN fieldDef text)
  begin
    IF NOT EXISTS (
        SELECT * FROM information_schema.COLUMNS
        WHERE column_name=fieldName
              and table_name=tableName
              and table_schema=dbName
    )
    THEN
      set @ddl=CONCAT('ALTER TABLE ',dbName,'.',tableName,
                      ' ADD COLUMN ',fieldName,' ',fieldDef);
      prepare stmt from @ddl;
      execute stmt;
    END IF;
  end;
# $$
#
# DELIMITER ';'














