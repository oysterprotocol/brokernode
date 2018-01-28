<?php

namespace App\Console\Commands;

use App\Clients\BrokerNode;
use App\DataMap;
use Carbon\Carbon;
use DB;
use Illuminate\Console\Command;

class CheckChunkStatus extends Command
{
    public static $chunkOut = '';

    const HOOKNODE_TIMEOUT_THRESHOLD_MINUTES = 20;

    protected $signature = 'CheckChunkStatus:checkStatus';
    protected $description =
        'Polls the status of chunks that have been sent to hook nodes';

    /**
     * Execute the console command.
     */
    public static function handle()
    {
        self::$chunkOut .= "At handle start. \n\n";
        echo self::$chunkOut;

        $thresholdTime = Carbon::now()
            ->subMinutes(self::HOOKNODE_TIMEOUT_THRESHOLD_MINUTES)
            ->toDateTimeString();

        self::$chunkOut .= "Threshold time: " . $thresholdTime . "\n\n";

        self::updateUnverifiedDatamaps($thresholdTime);
        self::updateTimedoutDatamaps($thresholdTime);
        self::purgeCompletedSessions();
    }

    /**
     * Private
     * */
    private static function updateUnverifiedDatamaps($thresholdTime)
    {
        self::$chunkOut .= "In updateUnverifiedDatamaps\n\n";

        $datamaps_unverified =
            DataMap::where('status', DataMap::status['unverified'])
                ->where('updated_at', '>=', $thresholdTime)
                ->get();

        $attached_datamaps = array_filter($datamaps_unverified->toArray(), function ($dmap) {
            // TODO: Make these concurrent.
            $req = (object)[
                "address" => $dmap["address"],
                "message" => $dmap["message"],
                "trunkTransaction" => $dmap["trunkTransaction"],
                "branchTransaction" => $dmap["branchTransaction"],
                "chunk_idx" => $dmap["chunk_idx"],
                "hooknode_id" => $dmap["hooknode_id"],
            ];

            $is_attached = BrokerNode::verifyChunkMessageMatchesRecord($req);
            //$is_attached = !BrokerNode::dataNeedsAttaching($req);
            /*
             * replace with 'verifyChunkMatchesRecord' if we also want to check
             * branch and trunk match the record.
             *
             * verifyChunkMessageMatchesRecord and verifyChunkMatchesRecord both
             * check tangle for the address and makes sure the message matches,
             * verifyChunkMatchesRecord also checks trunk and branch.
             */
            return $is_attached;
        });

        self::$chunkOut .= "datamaps_unverified: " . count($datamaps_unverified) . "\n\n";
        self::$chunkOut .= "attached_datamaps: " . count($attached_datamaps) . "\n\n";

        unset($datamaps_unverified); // Purges unused memory.

        $attached_ids = array_map(function ($dmap) {
            return $dmap["id"];
        }, $attached_datamaps);

        self::$chunkOut .= "attached_ids: " . count($attached_ids) . "\n\n";

        // Mass Update DB.
        DataMap::whereIn('id', $attached_ids)->update(['status' => DataMap::status['complete']]);

        self::updateHooknodeReputations($attached_datamaps);
    }

    private static function updateTimedoutDatamaps($thresholdTime)
    {

        self::$chunkOut .= "in updateTimedoutDatamaps.\n\n";

        $datamaps_timedout =
            DataMap::where('status', DataMap::status['unverified'])
                ->where('updated_at', '<', $thresholdTime)
                ->get();

        self::$chunkOut .= "datamaps_timedout: " . count($datamaps_timedout) . "\n\n";

        // TODO: Retry with another hooknode.
        return true; // placeholder.
    }

    private static function updateHooknodeReputations($attached_datamaps)
    {
        self::$chunkOut .= "in updateHooknodeReputations stub. \n\n";
        // TODO: Increment hooknode reputations for $attached_datamaps.
        return true; // placeholder.
    }

    public static function purgeCompletedSessions()
    {
        self::$chunkOut .= "in purgeCompletedSessions.\n\n";

        $not_complete_gen_hash = DB::table('data_maps')
            ->where('status', '<>', DataMap::status['complete'])
            ->select('genesis_hash', DB::raw('COUNT(genesis_hash) as not_completed'))
            ->groupBy('genesis_hash')
            ->pluck('genesis_hash');

        $completed_gen_hash = DB::table('upload_sessions')
            ->whereNotIn('genesis_hash', $not_complete_gen_hash)
            ->pluck('genesis_hash');

        self::$chunkOut .= "not_complete_gen_hash: " . count($not_complete_gen_hash) . "\n\n";
        self::$chunkOut .= "completed_gen_hash: " . count($completed_gen_hash) . "\n\n";

        DB::transaction(function () use ($completed_gen_hash) {
            DB::table('data_maps')
                ->whereIn('genesis_hash', $completed_gen_hash)
                ->delete();

            DB::table('upload_sessions')
                ->whereIn('genesis_hash', $completed_gen_hash)
                ->delete();

            self::$chunkOut .= "In purge callback\n\n";
        });

        echo self::$chunkOut;
    }
}
