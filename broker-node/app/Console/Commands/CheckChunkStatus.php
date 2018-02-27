<?php

namespace App\Console\Commands;

use Segment;
use App\Clients\BrokerNode;
use App\DataMap;
use App\HookNode;
use Carbon\Carbon;
use DB;
use Illuminate\Console\Command;

class CheckChunkStatus extends Command
{
    const HOOKNODE_TIMEOUT_THRESHOLD_MINUTES = 1;
    const CHUNKS_PER_REQUEST = 10;

    protected $signature = 'CheckChunkStatus:checkStatus';
    protected $description =
        'Polls the status of chunks that have been sent to hook nodes';

    private static $SegmentStarted = null;

    private static function initSegment()
    {
        if (is_null(self::$SegmentStarted )) {
            Segment::init("SrQ0wxvc7jp2XDjZiEJTrkLAo4FC2XdD");
            self::$SegmentStarted = true;
        }
    }

    /**
     * Execute the console command.
     */
    public static function handle()
    {
        //self::initSegment();

        $thresholdTime = Carbon::now()
            ->subMinutes(self::HOOKNODE_TIMEOUT_THRESHOLD_MINUTES)
            ->toDateTimeString();

        self::processUnassignedChunks($thresholdTime);
        self::updateUnverifiedDatamaps($thresholdTime);
        self::updateTimedoutDatamaps($thresholdTime);
        self::purgeCompletedSessions();
    }

    /**
     * Private
     * */

    private static function processUnassignedChunks($thresholdTime)
    {
        $datamaps_unassigned =
            DataMap::where('status', DataMap::status['unassigned'])
                ->where('message', '<>', null)
                ->where('updated_at', '>=', $thresholdTime)
                ->get()
                ->toArray();

        if (count($datamaps_unassigned)) {

            $chunkedChunkArrays = array_chunk($datamaps_unassigned, self::CHUNKS_PER_REQUEST);

            foreach ($chunkedChunkArrays as $chunkedChunkArray) {

                BrokerNode::processChunks($chunkedChunkArray, true);
            }
        }
    }

    private static function updateUnverifiedDatamaps($thresholdTime)
    {
        $datamaps_unverified =
            DataMap::where('status', DataMap::status['unverified'])
                ->where('message', '<>', null)
                ->where('updated_at', '>=', $thresholdTime)
                ->get()
                ->toArray();

        if (count($datamaps_unverified)) {

            $chunkedChunkArrays = array_chunk($datamaps_unverified, self::CHUNKS_PER_REQUEST);

            foreach ($chunkedChunkArrays as $chunkedChunkArray) {

                $filteredChunks = BrokerNode::verifyChunkMessagesMatchRecord($chunkedChunkArray);

                /*
                replace with 'verifyChunksMatchRecord' if we also want to check
                branch and trunk match the record.

                verifyChunkMessagesMatchRecord and verifyChunksMatchRecord both
                check tangle for the address and makes sure the message matches,
                verifyChunksMatchRecord also checks trunk and branch.
                */

                unset($chunkedChunkArray); // Purges unused memory.

                if (count($filteredChunks->matchesTangle)) {

                    $attached_ids = [];

                    array_walk($filteredChunks->matchesTangle, function ($dmap) use ($attached_ids) {
                        //record event
                        Segment::track([
                            "userId" => $_SERVER['SERVER_ADDR'],
                            "event" => "chunk_matches_tangle",
                            "properties" => ["hooknode_url" => $dmap->hooknode_id,
                                "chunk_idx" => $dmap->chunk_idx
                            ]
                        ]);
                        $attached_ids[] = $dmap->id;
                    });

                    // Mass Update DB.
                    DataMap::whereIn('id', $attached_ids)
                        ->update(['status' => DataMap::status['complete']]);

                    self::incrementHooknodeReputations($filteredChunks->matchesTangle);
                }

                if (count($filteredChunks->doesNotMatchTangle)) {

                    self::decrementHooknodeReputations($filteredChunks->doesNotMatchTangle);

                    $not_matching_ids = [];

                    array_walk($filteredChunks->doesNotMatchTangle, function ($dmap) use ($not_matching_ids) {
                        //record event
                        Segment::track([
                            "userId" => $_SERVER['SERVER_ADDR'],
                            "event" => "chunk_does_not_match_tangle",
                            "properties" => [
                                "hooknode_url" => $dmap->hooknode_id,
                                "chunk_idx" => $dmap->chunk_idx
                            ]
                        ]);
                        $not_matching_ids[] = $dmap->id;
                    });

                    DataMap::whereIn('id', $not_matching_ids)
                        ->update([
                            'status' => DataMap::status['unassigned'],
                            'hooknode_id' => null,
                            'branchTransaction' => null,
                            'trunkTransaction' => null]);

                    BrokerNode::processChunks($filteredChunks->doesNotMatchTangle, true);
                }
            }
        }
    }

    private static function updateTimedoutDatamaps($thresholdTime)
    {
        $datamaps_timedout =
            DataMap::where('updated_at', '<', $thresholdTime)
                ->where('status', '<>', DataMap::status['complete'])
                ->where('message', '<>', null)
                ->get()
                ->toArray();

        if (count($datamaps_timedout)) {

            $chunkedChunkArrays = array_chunk($datamaps_timedout, self::CHUNKS_PER_REQUEST);

            foreach ($chunkedChunkArrays as $chunkedChunkArray) {

                self::decrementHooknodeReputations($chunkedChunkArray);

                $timed_out_ids = [];

                array_walk($chunkedChunkArray, function ($dmap) use ($timed_out_ids) {
                    //record event
                    Segment::track([
                        "userId" => $_SERVER['SERVER_ADDR'],
                        "event" => "resending_chunk",
                        "properties" => [
                            "hooknode_url" => $dmap['hooknode_id'],
                            "chunk_idx" => $dmap['chunk_idx']
                        ]
                    ]);
                    $timed_out_ids[] = $dmap['id'];
                });

                // Mass Update DB.
                DataMap::whereIn('id', $timed_out_ids)
                    ->update([
                        'status' => DataMap::status['unassigned'],
                        'hooknode_id' => null,
                        'branchTransaction' => null,
                        'trunkTransaction' => null]);

                BrokerNode::processChunks($chunkedChunkArray);
            }
        }
    }

    private static function incrementHooknodeReputations($datamaps)
    {
        $unique_hooks = self::getUniqueHooks($datamaps);

        foreach ($unique_hooks as $hook) {
            //record event
            Segment::track([
                "userId" => $_SERVER['SERVER_ADDR'],
                "event" => "hooknode_score_increment",
                "properties" => [
                    "hooknode_url" => $hook
                ]
            ]);
            HookNode::incrementScore($hook);
        }
    }

    private static function decrementHooknodeReputations($datamaps)
    {
        $unique_hooks = self::getUniqueHooks($datamaps);

        foreach ($unique_hooks as $hook) {
            //record event
            Segment::track([
                "userId" => $_SERVER['SERVER_ADDR'],
                "event" => "hooknode_score_decrement",
                "properties" => [
                    "hooknode_url" => $hook
                ]
            ]);
            HookNode::decrementScore($hook);
        }
    }

    private static function getUniqueHooks($datamaps)
    {
        $hooknode_ids = array();

        foreach ($datamaps as $datamap) {
            $hooknode_ids[] = is_object($datamap) ?
                $datamap->hooknode_id : $datamap['hooknode_id'];
        }

        // array_filter removes empty values when not provided a callback
        return array_unique(array_filter($hooknode_ids));
    }

    public static function purgeCompletedSessions()
    {
        $not_complete_gen_hash = DB::table('data_maps')
            ->where('status', '<>', DataMap::status['complete'])
            ->select('genesis_hash', DB::raw('COUNT(genesis_hash) as not_completed'))
            ->groupBy('genesis_hash')
            ->pluck('genesis_hash');

        $completed_gen_hash = DB::table('upload_sessions')
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
