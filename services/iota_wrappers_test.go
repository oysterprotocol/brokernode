package services_test

import (
	"fmt"
	"github.com/oysterprotocol/brokernode/models"
	"github.com/oysterprotocol/brokernode/services"
	"testing"
	"time"
)

func Test_Init(t *testing.T) {

	//it should have determined the number of PoW processes
	if services.PowProcs == 0 {
		t.Fatalf("init should have determined the PowProcs")
	}

	if len(services.Channel) != services.PowProcs {
		t.Fatalf("init should have made 1 channel for each PowProc")
	}

	channels := []models.ChunkChannel{}

	models.DB.RawQuery("Select * from chunk_channels").All(&channels)

	for _, channel := range channels {
		if _, ok := services.Channel[channel.ChannelID]; !ok {
			t.Fatalf("after init, for every channel in chunk_channels there should be a corresponding "+
				"channel in services.Channel with the same ChannelID, but ChannelID %s is missing from "+
				"services.Channel or there is an extra channel in the DB.", channel.ChannelID)
		}
	}
}

func Test_SetEstimatedReadyTime(t *testing.T) {

	chunkTracker := []services.ChunkTracker{
		{ElapsedTime: 1 * time.Minute, ChunkCount: 6},
		{ElapsedTime: 1 * time.Minute, ChunkCount: 5},
	}
	//this will yield an average time of 11 seconds per chunk

	channels := []models.ChunkChannel{}
	models.DB.RawQuery("Select * from chunk_channels").All(&channels)

	*(services.Channel[channels[0].ChannelID].ChunkTrackers) = chunkTracker

	currentTime := time.Now()

	channels[0].EstReadyTime = services.SetEstimatedReadyTime(
		services.Channel[channels[0].ChannelID],
		3)

	models.DB.ValidateAndSave(&channels[0])

	channel := models.ChunkChannel{}

	_ = models.DB.Where("channel_id = ?", channels[0].ChannelID).First(&channel)

	result := channel.EstReadyTime.Sub(currentTime)

	// With an average time of 11 seconds per chunk and 3 chunks passed in, we should expect
	// EstReadyTime to be set about 33 seconds in the future
	if result > time.Duration(35*time.Second) || result < time.Duration(31*time.Second) {
		fmt.Println(result)
		t.Fatalf("SetEstimatedReadyTime:  the average time per chunk was 11 seconds, so " +
			"for 3 chunks our EstReadyTime should have been roughly 33 seconds from now")
	}
}

func Test_TrackProcessingTime(t *testing.T) {

	startTime := time.Now().Add(-1 * time.Minute)

	channels := []models.ChunkChannel{}

	models.DB.RawQuery("Select * from chunk_channels").All(&channels)

	powChannel := services.Channel[channels[0].ChannelID]

	initialLastChunkRecord := ((*(powChannel.ChunkTrackers))[len(*powChannel.ChunkTrackers)-1])

	initialPoWFrequency := services.PoWFrequency.Frequency

	services.TrackProcessingTime(startTime, 10, &powChannel)

	newLastChunkRecord := ((*(powChannel.ChunkTrackers))[len(*powChannel.ChunkTrackers)-1])

	// check that we have added a new record
	if newLastChunkRecord == initialLastChunkRecord {
		t.Fatalf("TrackProcessingTime:  should have added a new chunk record to the end of " +
			"ChunkTrackers")
	}

	// call the method more than 10 times
	for i := 0; i < 15; i++ {
		services.TrackProcessingTime(startTime, 10, &powChannel)
	}

	// check that there are only 10 records
	if len(*powChannel.ChunkTrackers) != 10 {
		t.Fatalf("TrackProcessingTime:  only supposed to hold the last 10 records")
	}

	if initialPoWFrequency == services.GetProcessingFrequency() {
		t.Fatalf("TrackProcessingTime:  PoWFrequency.Frequency should have changed")
	}
}
