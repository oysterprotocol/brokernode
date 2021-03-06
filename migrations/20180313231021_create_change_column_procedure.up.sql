drop procedure if exists ChangeColumnIfExists;
# DELIMITER '$$'

create procedure ChangeColumnIfExists(
  IN dbName tinytext,
  IN tableName tinytext,
  IN fieldName tinytext,
  IN fieldDef text)
  begin
    IF EXISTS (
        SELECT * FROM information_schema.COLUMNS
        WHERE column_name=fieldName
              and table_name=tableName
              and table_schema=dbName
    )
    THEN
      set @ddl=CONCAT('ALTER TABLE ',dbName,'.',tableName,
                      ' MODIFY COLUMN ',fieldName,' ',fieldDef);
      prepare stmt from @ddl;
      execute stmt;
    END IF;
  end;
# $$
#
# DELIMITER ';'