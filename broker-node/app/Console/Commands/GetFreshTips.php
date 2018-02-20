<?php

namespace App\Console\Commands;

use App\Clients\BrokerNode;
use App\DataMap;
use App\Tips;
use App\HookNode;
use Carbon\Carbon;
use DB;
use Illuminate\Console\Command;

class GetFreshTips extends Command
{
    const MINS_TO_RUN_PROCESS = 1;
    const TIPS_THRESHOLD_MINUTES = 1.5;
    const NUM_HOOKS_TO_QUERY = 2;

    protected $signature = 'GetFreshTips:getTips';
    protected $description =
        'Gets a batch of tips that need approving';

    /**
     * Execute the console command.
     */
    public static function handle()
    {
        echo "Running GetFreshTips at: " . Carbon::now() . "\n";

        $tipsThresholdTime = Carbon::now()
            ->subMinutes(self::TIPS_THRESHOLD_MINUTES)
            ->toDateTimeString();

        $processThresholdTime = Carbon::now()
            ->addMinutes(self::MINS_TO_RUN_PROCESS)
            ->toDateTimeString();

        self::getFreshTipsFromSelf();
        self::getFreshTipsFromHookNodes($processThresholdTime);
        self::purgeOldTips($tipsThresholdTime);
    }

    /**
     * Private
     * */

    private static function getFreshTipsFromSelf()
    {
        //logic to getBulkTransactionsToApprove

        /*TODO: remove all this*/
        $arrayOfTips = array();
        //put actual logic in here

        Tips::addTips($arrayOfTips);
    }

    private static function getFreshTipsFromHookNodes($thresholdTime)
    {
        $numHookNodesQueried = 0;

        while ($thresholdTime->gt(Carbon::now()) && $numHookNodesQueried <= self::NUM_HOOKS_TO_QUERY) {

            /*TODO: remove all this*/

            $ready = false;

            while ($ready == false) {
                [$ready, $next_node] = HookNode::getNextReadyNode();
            }

            $node_address = $next_node->ip_address;

            //put actual logic here to get the tips
            $arrayOfTips = array();
            //put actual logic in here

            HookNode::setTimeOfLastContact($node_address);
            HookNode::incrementScore($node_address);

            $numHookNodesQueried++;
            Tips::addTips($arrayOfTips, $node_address);
        }
    }

    private static function purgeOldTips($tipsThresholdTime)
    {
        Tips::where('updated_at', '<', $tipsThresholdTime)
            ->delete();
    }
}
