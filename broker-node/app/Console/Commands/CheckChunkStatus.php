<?php

namespace App\Console\Commands;

use App\Clients\BrokerNode;
use App\DataMap;
use App\HookNode;
use Carbon\Carbon;
use DB;
use Illuminate\Console\Command;

class CheckChunkStatus extends Command
{
    const HOOKNODE_TIMEOUT_THRESHOLD_MINUTES = 2;

    protected $signature = 'CheckChunkStatus:checkStatus';
    protected $description =
        'Polls the status of chunks that have been sent to hook nodes';

    /**
     * Execute the console command.
     */
    public static function handle()
    {
        $thresholdTime = Carbon::now()
            ->subMinutes(self::HOOKNODE_TIMEOUT_THRESHOLD_MINUTES)
            ->toDateTimeString();

        // self::processUnassignedChunks();
        self::updateUnverifiedDatamaps($thresholdTime);
        self::updateTimedoutDatamaps($thresholdTime);
        self::purgeCompletedSessions();
    }

    /**
     * Private
     * */

    private static function processUnassignedChunks()
    {
        $unassigned_datamaps = DataMap::getUnassigned()->all();
        foreach ($unassigned_datamaps as &$dmap) { // TODO: Concurrent.
            $dmap->processChunk();
        }
    }

    private static function updateUnverifiedDatamaps($thresholdTime)
    {
        $datamaps_unverified =
            DataMap::where('status', DataMap::status['unverified'])
                ->where('updated_at', '>=', $thresholdTime)
                ->get()
                ->toArray();

        if (count($datamaps_unverified)) {

            $filteredChunks = BrokerNode::verifyChunkMessagesMatchRecord($datamaps_unverified);

            /*
            replace with 'verifyChunksMatchRecord' if we also want to check
            branch and trunk match the record.

            verifyChunkMessagesMatchRecord and verifyChunksMatchRecord both
            check tangle for the address and makes sure the message matches,
            verifyChunksMatchRecord also checks trunk and branch.
            */

            unset($datamaps_unverified); // Purges unused memory.

            if (count($filteredChunks->matchesTangle)) {
                $attached_ids = array_map(function ($dmap) {
                    return $dmap->id;
                }, $filteredChunks->matchesTangle);

                // Mass Update DB.
                DataMap::whereIn('id', $attached_ids)
                    ->update(['status' => DataMap::status['complete']]);

                self::incrementHooknodeReputations($filteredChunks->matchesTangle);
            }

            if (count($filteredChunks->doesNotMatchTangle)) {
                $not_matching_ids = array_map(function ($dmap) {
                    return $dmap->id;
                }, $filteredChunks->doesNotMatchTangle);

                // Mass Update DB.
                DataMap::whereIn('id', $not_matching_ids)
                    ->update(['status' => DataMap::status['unassigned']]);

                BrokerNode::processChunks($filteredChunks->doesNotMatchTangle);

                self::decrementHooknodeReputations($filteredChunks->doesNotMatchTangle);
            }
        }
    }

    private static function updateTimedoutDatamaps($thresholdTime)
    {
        $datamaps_timedout =
            DataMap::where('status', DataMap::status['unverified'])
                ->where('updated_at', '<', $thresholdTime)
                ->get()
                ->toArray();

        if (count($datamaps_timedout)) {

            self::decrementHooknodeReputations($datamaps_timedout);

            BrokerNode::processChunks($datamaps_timedout);
            //will this work?

            // TODO: Retry with another hooknode.

            // NOT FOR TESTNET.
            // $datamaps_timedout_query->update([
            //     'status' => DataMap::status['unassigned'],
            //     'hooknode_id' => null,
            //     'branchTransaction' => null,
            //     'trunkTransaction' => null,
            // ]);
            // // Note: DB and in memory model are now out of sync, but it should be ok...
            // foreach ($datamaps_timedout as &$dmap) { // TODO: Concurrent.
            //     $dmap->processChunk();
            // }
        }
    }

    private static function incrementHooknodeReputations($datamaps)
    {
        $unique_hooks = self::getUniqueHooks($datamaps);

        foreach ($unique_hooks as $hook) {
            HookNode::incrementScore($hook);
        }
    }

    private static function decrementHooknodeReputations($datamaps)
    {
        $unique_hooks = self::getUniqueHooks($datamaps);

        foreach ($unique_hooks as $hook) {
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

        return array_unique($hooknode_ids);
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
