<?php

namespace App\Console\Commands;

use \Exception;
use App\Clients\requests\IriData;
use App\Clients\requests\IriWrapper;
use App\Tips;
use Carbon\Carbon;
use Illuminate\Console\Command;

class GetFreshTips extends Command
{
    const TIPS_THRESHOLD_SECONDS = 5;
    const PROCESS_RUN_TIME_SECONDS = 295;

    const TIPS_QUANTITY = 50;

    protected $signature = 'GetFreshTips:getTips';
    protected $description =
        'Gets a batch of tips that need approving';

    /**
     * Execute the console command.
     */
    public static function handle()
    {
        $tipsThresholdTime = Carbon::now()
            ->subSeconds(self::TIPS_THRESHOLD_SECONDS)
            ->toDateTimeString();

        $processStopTime = Carbon::now()
            ->addSeconds(self::PROCESS_RUN_TIME_SECONDS);

        while (Carbon::now()->lt($processStopTime)) {
            self::getFreshTipsFromSelf();
            self::purgeOldTips($tipsThresholdTime);
            sleep(5);
        }
    }

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

    private static function purgeOldTips($tipsThresholdTime)
    {
        Tips::where('updated_at', '<', $tipsThresholdTime)
            ->delete();
    }
}
