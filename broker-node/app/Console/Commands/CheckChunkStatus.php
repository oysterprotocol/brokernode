<?php

namespace App\Console\Commands;

use App\Clients\BrokerNode;
use App\DataMap;
use Carbon\Carbon;
use DB;
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
        self::purgeCompletedSessions();
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
                "address" => $dmap["hash"],
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

        self::updateHooknodeReputations($attached_datamaps);
    }

    private function updateTimedoutDatamaps($thresholdTime) {
        $datamaps_timedout =
            DataMap::where('status', 'pending')
                ->where('updated_at', '<', $thresholdTime)
                ->get();

        // TODO: Retry with another hooknode.
        return true; // placeholder.
    }

    private static function updateHooknodeReputations($attached_datamaps) {
        // TODO: Increment hooknode reputations for $attached_datamaps.
        return true; // placeholder.
    }

    public static function purgeCompletedSessions() {
        $not_complete_gen_hash = DB::table('data_maps')
            ->where('status', '<>', 'complete')
            ->select('genesis_hash', DB::raw('COUNT(genesis_hash) as not_completed'))
            ->groupBy('genesis_hash')
            ->pluck('genesis_hash');

        $completed_gen_hash =  DB::table('upload_sessions')
            ->whereNotIn('genesis_hash', $not_complete_gen_hash)
            ->pluck('genesis_hash');

        DB::transaction(function () use ($completed_gen_hash) {
            DB::table('data_maps')
                ->whereIn('genesis_hash', $completed_gen_hash)
                ->delete();

            DB::table('upload_sessions')
                ->whereIn('genesis_hash', $completed_gen_hash)
                ->delete();
        });
    }
}
