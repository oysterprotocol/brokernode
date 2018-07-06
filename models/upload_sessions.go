func (u *UploadSession) BulkMarkDataMapsAsUnassigned() error {
	var err error
	for i := 0; i < oyster_utils.MAX_NUMBER_OF_SQL_RETRY; i++ {
		err = DB.RawQuery("UPDATE data_maps SET status = ? "+
			"WHERE id IN (SELECT id FROM data_maps WHERE genesis_hash = ? AND status = ? AND message != ? AND msg_status = ?)",
			Unassigned,
			u.GenesisHash,
			Pending,
			DataMap{}.Message,
			MsgStatusUnmigrated).All(&[]DataMap{})
		if err == nil {
			break
		}
	}
	oyster_utils.LogIfError(err, map[string]interface{}{"MaxRetry": oyster_utils.MAX_NUMBER_OF_SQL_RETRY})

	err = nil
	for i := 0; i < oyster_utils.MAX_NUMBER_OF_SQL_RETRY; i++ {
		err = DB.RawQuery("UPDATE data_maps SET status = ? "+
			"WHERE id IN (SELECT id FROM data_maps WHERE genesis_hash = ? AND status = ? AND msg_status = ?)",
			Unassigned,
			u.GenesisHash,
			Pending,
			MsgStatusUploaded).All(&[]DataMap{})
		if err == nil {
			break
		}
	}
	oyster_utils.LogIfError(err, map[string]interface{}{"MaxRetry": oyster_utils.MAX_NUMBER_OF_SQL_RETRY})

	return err
}
