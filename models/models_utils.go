package models

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/oysterprotocol/brokernode/utils"
)

/*BatchUpsert updates a table and overwrite its current the values of column.*/
func BatchUpsert(tableName string, serializeValues []string, serializedColumnNames string, onConflictColumnsNames []string) error {
	numOfBatchRequest := int(math.Ceil(float64(len(serializeValues)) / float64(oyster_utils.SQL_BATCH_SIZE)))

	var updatedColumns []string
	for _, v := range onConflictColumnsNames {
		updatedColumns = append(updatedColumns, fmt.Sprintf("%s = VALUES(%s)", v, v))
	}
	serializedUpdatedColumnName := strings.Join(updatedColumns, oyster_utils.COLUMNS_SEPARATOR)

	// Batch Update data_maps table.
	remainder := len(serializeValues)
	for i := 0; i < numOfBatchRequest; i++ {
		lower := i * oyster_utils.SQL_BATCH_SIZE
		upper := i*oyster_utils.SQL_BATCH_SIZE + int(math.Min(float64(remainder), oyster_utils.SQL_BATCH_SIZE))

		upsertedValues := serializeValues[lower:upper]
		for i = 0; i < len(upsertedValues); i++ {
			upsertedValues[i] = fmt.Sprintf("(%s)", upsertedValues[i])
		}

		// Do an insert operation and dup by primary key.
		var rawQuery string
		values := strings.Join(upsertedValues, oyster_utils.COLUMNS_SEPARATOR)
		if len(serializedUpdatedColumnName) > 0 {
			rawQuery = fmt.Sprintf(`INSERT INTO %s(%s) VALUES %s ON DUPLICATE KEY UPDATE %s`,
				tableName, serializedColumnNames, values, serializedUpdatedColumnName)
		} else {
			rawQuery = fmt.Sprintf(`INSERT INTO %s(%s) VALUES %s`, tableName, serializedColumnNames, values)
		}

		fmt.Println("hello world")
		fmt.Println(rawQuery)

		err := DB.RawQuery(rawQuery).Exec()
		fmt.Println(err)
		retryCount := oyster_utils.MAX_NUMBER_OF_SQL_RETRY
		for err != nil && retryCount > 0 {
			time.Sleep(300 * time.Millisecond)
			err = DB.RawQuery(rawQuery).Exec()
			retryCount = retryCount - 1
		}
		if err != nil {
			oyster_utils.LogIfError(err, nil)
			return err
		}

		remainder = remainder - oyster_utils.SQL_BATCH_SIZE
	}
	return nil
}
