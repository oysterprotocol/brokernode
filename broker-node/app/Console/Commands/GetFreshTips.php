<?php

namespace App\Console\Commands;

use \Exception;
use App\Clients\BrokerNode;
use App\Clients\requests\IriData;
use App\Clients\requests\IriWrapper;
use App\DataMap;
use \stdClass;
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
    const TIPS_QUANTITY = 100;

    protected $signature = 'GetFreshTips:getTips';
    protected $description =
        'Gets a batch of tips that need approving';

    /**
     * Execute the console command.
     */
    public static function handle()
    {
        $tipsThresholdTime = Carbon::now()
            ->subMinutes(self::TIPS_THRESHOLD_MINUTES)
            ->toDateTimeString();

        $processThresholdTime = Carbon::now()
            ->addMinutes(self::MINS_TO_RUN_PROCESS)
            ->toDateTimeString();

        self::getFreshTipsFromSelf();
        //self::getFreshTipsFromHookNodes($processThresholdTime);
        self::purgeOldTips($tipsThresholdTime);

    }

    /**
     * Private
     * */

    private static function getFreshTipsFromSelf()
    {
        $command = new \stdClass();
        $command->command = "getBulkTransactionsToApprove";
        $command->depth = IriData::$depthToSearchForTxs;
        $command->quantity = self::TIPS_QUANTITY;

        $result = IriWrapper::makeRequest($command);

        if (!is_null($result) && property_exists($result, 'transactions')) {
            Tips::addTips($result->transactions);
        } else {
            $error = '';
            foreach ($result as $key => $value) {
                if (is_array($value)) {
                    $error .= $key . ": \n" . implode("\n", $value) . "\n\n";
                } else {
                    $error .= $key . ': ' . $value . "\n";
                }
                $error .= "\n";
            }

            throw new \Exception('getBulkTransactionsToApprove failed!' . $error);
        }
    }

    private static function getFreshTipsFromHookNodes($thresholdTime)
    {
        $numHookNodesQueried = 0;

        while ($thresholdTime->gt(Carbon::now()) && $numHookNodesQueried <= self::NUM_HOOKS_TO_QUERY) {

            /*TODO: remove all this*/

            $next_node = HookNode::getNextReadyNode();

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
