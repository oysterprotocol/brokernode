drop procedure if exists AddConstraintUnlessExists;
# DELIMITER '$$'

create procedure AddConstraintUnlessExists(
  IN dbName         tinytext,
  IN tableName      tinytext,
  IN constraintName text,
  IN constraintType text,
  IN constraintDef  text)
  begin
    IF NOT EXISTS(
        SELECT *
        FROM information_schema.TABLE_CONSTRAINTS
        WHERE table_name = tableName
              and constraint_name = constraintName
              and constraint_type = constraintType
              and table_schema = dbName
    )
    THEN
      set @ddl = CONCAT('ALTER TABLE ', dbName, '.', tableName,
                        ' ADD CONSTRAINT ', constraintName, ' ', constraintDef);
      prepare stmt from @ddl;
      execute stmt;
    END IF;
  end;
