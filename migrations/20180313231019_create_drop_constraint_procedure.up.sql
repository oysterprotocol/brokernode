drop procedure if exists DropConstraintIfExists;
# DELIMITER '$$'

create procedure DropConstraintIfExists(
  IN dbName         tinytext,
  IN tableName      tinytext,
  IN constraintName text,
  IN constraintType text)
  begin
    IF EXISTS(
        SELECT *
        FROM information_schema.TABLE_CONSTRAINTS
        WHERE table_name = tableName
              and constraint_name = constraintName
              and constraint_type = constraintType
              and table_schema = dbName
    )
    THEN
      set @ddl = CONCAT('ALTER TABLE ', dbName, '.', tableName,
                        ' DROP KEY ', constraintName);
      prepare stmt from @ddl;
      execute stmt;
    END IF;
  end;