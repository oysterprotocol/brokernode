<?php

namespace App\Console\Commands;

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
        $thresholdTime =
            Carbon::now()->subMinutes(self::HOOKNODE_TIMEOUT_THRESHOLD_MINUTES)
                ->toDateTimeString();

        /**
         * Check pending datamaps.
         */

        $datamaps_pending =
            DataMap::where('status', 'pending')
                ->where('updated_at', '>=', $thresholdTime)
                ->get();

        $attached_datamaps = array_filter($datamaps_pending->toArray(), function($dmap) {
            // TODO: Check status on tangle.
            // TODO: Make these concurrent.
            $is_attached = true; // placeholder.
            return $is_attached;
        });
        $attached_ids = array_map(function($dmap) {
            return $dmap["id"];
        }, $attached_datamaps);

        // Mass Update DB.
        DataMap::whereIn('id', $attached_ids)->update(['status', 'complete']);

        // TODO: Increment hooknode reputations for $attached_datamaps.

        unset($datamaps_pending); // Purges unused memory.

        /**
         * Retry timedout datamaps.
         */
        $datamaps_timedout =
            DataMap::where('status', 'pending')
                ->where('updated_at', '<', $thresholdTime)
                ->get();

        // TODO: Retry with another hooknode.
    }
}
