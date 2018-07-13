ALTER TABLE `treasures`
ADD msg_id VARCHAR (255) DEFAULT NULL;

ALTER TABLE `completed_data_maps`
ADD msg_id VARCHAR (255) DEFAULT NULL;

ALTER TABLE `data_maps`
ADD msg_id VARCHAR (255) NOT NULL;
