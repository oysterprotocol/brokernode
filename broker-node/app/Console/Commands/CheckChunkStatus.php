<?php

namespace App\Console\Commands;

use App\Clients\BrokerNode;
use App\DataMap;
use Carbon\Carbon;
use Illuminate\Console\Command;

class CheckChunkStatus extends Command
{
    const HOOKNODE_TIMEOUT_THRESHOLD_MINUTES = 20;

    protected $signature = 'CheckChunkStatus:checkStatus';
    protected $description =
        'Polls the status of chunks that have been sent to hook nodes';

    /**
     * Execute the console command.
     */
    public function handle() {
        $thresholdTime = Carbon::now()
            ->subMinutes(self::HOOKNODE_TIMEOUT_THRESHOLD_MINUTES)
            ->toDateTimeString();

        self::updatePendingDatamaps($thresholdTime);
        self::updateTimedoutDatamaps($thresholdTime);
    }

    /**
     * Private
     * */

    private function updatePendingDatamaps($thresholdTime) {
        $datamaps_pending =
            DataMap::where('status', 'pending')
                ->where('updated_at', '>=', $thresholdTime)
                ->get();

        $attached_datamaps = array_filter($datamaps_pending->toArray(), function($dmap) {
            // TODO: Make these concurrent.
            $req = (object)[
                "address" => $dmap->hash,
            ];
            $is_attached = !BrokerNode::dataNeedsAttaching($req);
            return $is_attached;
        });
        unset($datamaps_pending); // Purges unused memory.

        $attached_ids = array_map(function($dmap) {
            return $dmap["id"];
        }, $attached_datamaps);

        // Mass Update DB.
        DataMap::whereIn('id', $attached_ids)->update(['status' => 'complete']);

        updateHooknodeReputations($attached_datamaps);
    }

    private function updateTimedoutDatamaps($thresholdTime) {
        $datamaps_timedout =
            DataMap::where('status', 'pending')
                ->where('updated_at', '<', $thresholdTime)
                ->get();

        // TODO: Retry with another hooknode.
        return true; // placeholder.
    }

    private function updateHooknodeReputations($attached_datamaps) {
        // TODO: Increment hooknode reputations for $attached_datamaps.
        return true; // placeholder.
    }
}
