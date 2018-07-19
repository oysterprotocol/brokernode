drop procedure if exists DropKeyIfExists;
# DELIMITER '$$'

create procedure DropKeyIfExists(
  IN dbName         tinytext,
  IN tableName      tinytext,
  IN indexName text,
  IN nonUnique int)
  begin
    IF EXISTS(
        SELECT *
        FROM information_schema.STATISTICS
        WHERE table_name = tableName
              and index_name = indexName
              and non_unique = nonUnique
              and table_schema = dbName
    )
    THEN
      set @ddl = CONCAT('ALTER TABLE ', dbName, '.', tableName,
                        ' DROP KEY ', indexName);
      prepare stmt from @ddl;
      execute stmt;
    END IF;
  end;