drop procedure if exists DropColumnIfExists;
# DELIMITER '$$'

create procedure DropColumnIfExists(
  IN dbName tinytext,
  IN tableName tinytext,
  IN fieldName tinytext)
  begin
    IF EXISTS (
        SELECT * FROM information_schema.COLUMNS
        WHERE column_name=fieldName
              and table_name=tableName
              and table_schema=dbName
    )
    THEN
      set @ddl=CONCAT('ALTER TABLE ',dbName,'.',tableName,
                      ' DROP COLUMN ',fieldName);
      prepare stmt from @ddl;
      execute stmt;
    END IF;
  end;
# $$
#
# DELIMITER ';'