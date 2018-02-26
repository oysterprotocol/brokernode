<?php

namespace App;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Facades\DB;
use Carbon\Carbon;
use Webpatser\Uuid\Uuid;

class HookNode extends Model
{
    const SCORE_WEIGHT = 1;
    const TIME_WEIGHT = 1;

    const NUM_HOOKS_TO_BROADCAST = 10;

    /**
     * TODO: Make this a shared trait.
     *  Setup model event hooks
     */
    public static function boot()
    {
        parent::boot();
        self::creating(function ($model) {
            $model->id = (string)Uuid::generate(4);
        });
    }

    protected $table = 'hook_nodes';
    protected $fillable = ['ip_address', 'score', 'contacted_at'];
    public $incrementing = false; // UUID

    public static function insertNode($ip_address)
    {
        return self::create(['ip_address' => $ip_address]);
    }

    public static function getNextReadyNode($opts = [])
    {
        $score_weight = isset($opts['score_weight']) ? $opts['score_weight'] : self::SCORE_WEIGHT;
        $time_weight = isset($opts['time_weight']) ? $opts['time_weight'] : self::TIME_WEIGHT;
        $now = Carbon::now()->toDateTimeString();

        $nextNode = DB::table('hook_nodes')
            ->select(DB::raw("
                *,
                $score_weight * score
                    + $time_weight * TIMESTAMPDIFF(SECOND, contacted_at, '$now')
                    AS selection_score
            "))
            ->orderBy('selection_score', 'desc')
            ->first();

        self::setTimeOfLastContact($nextNode->ip_address);

        return $nextNode;
    }

    public static function getBroadcastingHookNodes($opts = [])
    {
        $score_weight = isset($opts['score_weight']) ? $opts['score_weight'] : self::SCORE_WEIGHT;
        $time_weight = isset($opts['time_weight']) ? $opts['time_weight'] : self::TIME_WEIGHT;
        $now = Carbon::now()->toDateTimeString();

        $broadcastingNodes = DB::table('hook_nodes')
            ->select(DB::raw("
                *,
                $score_weight * score
                    + $time_weight * TIMESTAMPDIFF(SECOND, contacted_at, '$now')
                    AS selection_score
            "))
            ->orderBy('selection_score', 'desc')
            ->take(self::NUM_HOOKS_TO_BROADCAST)
            ->get()
            ->toArray();

        // not updating 'contacted_at' here
        // don't want to hurt hook's chances of
        // getting picked for doing work that it
        // does not get rewarded for

        return $broadcastingNodes;
    }

    public static function incrementScore($ip_address)
    {
        DB::table('hook_nodes')
            ->where('ip_address', $ip_address)
            ->increment('score', 1);
    }

    public static function decrementScore($ip_address)
    {
        DB::table('hook_nodes')
            ->where('ip_address', $ip_address)
            ->decrement('score', 1);
    }

    public static function incrementChunksProcessed($ip_address, $chunks_count = 1)
    {
        DB::table('hook_nodes')
            ->where('ip_address', $ip_address)
            ->increment('chunks_processed_count', $chunks_count);
    }

    public static function setTimeOfLastContact($ip_address)
    {
        DB::table('hook_nodes')
            ->where('ip_address', $ip_address)
            ->update([
                'contacted_at' => Carbon::now()
            ]);
    }
}
