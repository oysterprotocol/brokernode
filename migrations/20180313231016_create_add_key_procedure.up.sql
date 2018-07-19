drop procedure if exists AddKeyUnlessExists;
# DELIMITER '$$'

create procedure AddKeyUnlessExists(
  IN dbName         tinytext,
  IN tableName      tinytext,
  IN indexName text,
  IN nonUnique int,
  IN indexDef text)
begin
    IF NOT EXISTS(
        SELECT *
        FROM information_schema.STATISTICS
        WHERE table_name = tableName
              and index_name = indexName
              and non_unique = nonUnique
              and table_schema = dbName
    )
    THEN
      set @ddl = CONCAT('ALTER TABLE ', dbName, '.', tableName,
                        ' ADD KEY ', indexName, ' ', indexDef);
      prepare stmt from @ddl;
      execute stmt;
    END IF;
  end;