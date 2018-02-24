<?php

namespace App\Console\Commands;

use \Exception;
use App\Clients\requests\IriData;
use App\Clients\requests\IriWrapper;
use App\Tips;
use App\HookNode;
use App\ChunkEvents;
use Carbon\Carbon;
use Illuminate\Console\Command;

class GetFreshTips extends Command
{
    const TIPS_THRESHOLD_MINUTES = .5;
    const TIPS_QUANTITY = 100;
    const PROCESS_RUN_TIME = .75;

    protected $signature = 'GetFreshTips:getTips';
    protected $description =
        'Gets a batch of tips that need approving';

    /**
     * Execute the console command.
     */
    public static function handle()
    {
        self::initEventRecord();

        $tipsThresholdTime = Carbon::now()
            ->subMinutes(self::TIPS_THRESHOLD_MINUTES)
            ->toDateTimeString();

        $processStopTime = Carbon::now()
            ->addMinutes(self::PROCESS_RUN_TIME)
            ->toDateTimeString();

        while (Carbon::now()->toDateTimeString() < $processStopTime) {
            self::$ChunkEventsRecord->addChunkEvent("re-running getFreshTips", "filler", "todo", "todo");
            self::getFreshTipsFromSelf();
            self::purgeOldTips($tipsThresholdTime);
        }
    }

    private static function initEventRecord()
    {
        if (is_null(self::$ChunkEventsRecord)) {
            self::$ChunkEventsRecord = new ChunkEvents();
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
