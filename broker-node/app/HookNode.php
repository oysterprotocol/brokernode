<?php

namespace App;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Facades\DB;
use Carbon\Carbon;
use Webpatser\Uuid\Uuid;

class HookNode extends Model
{
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
    protected $fillable = ['ip_address'];
    public $incrementing = false; // UUID

    public static function insertNode($ip_address)
    {
        return self::create(['ip_address' => $ip_address]);
    }

    public static function getNextReadyNode()
    {
        $nextNode =
            // ->where('status', "ready")  don't want to do this, we want to just ask
            // the hooknode its status
            HookNode::orderBy('score', 'desc')
                ->oldest('time_of_last_contact')
                ->first();

        if (self::isHookNodeAvailable($nextNode->ip_address)) {
            return [true, $nextNode];
        } else {
            return [false, null];
        }
    }

    private static function isHookNodeAvailable($ip_address)
    {
        self::setTimeOfLastContact($ip_address);

        // For this method we need to call the hooknode and ask it if it is
        // available for work.  I don't think that's implemented yet on the hooknodes,
        // so for now just giving a random 1 in 5 chance that the hooknode will say it is
        // busy.  Changed 'time_of_last_chunk' to 'time_of_last_contact'.  This
        // will set the timestamp regardless of whether we ultimately send the hook
        // a chunk or not, we don't want to keep asking the same hooknode over and
        // over if it is available.

        $randomChance = rand(1, 5);

        if ($randomChance == 5) {
            return false;
        } else {
            return true;
        }
    }

    public static function incrementScore($ip_address)
    {
        HookNode::where('ip_address', $ip_address)
            ->increment('score', 1);
    }

    public static function decrementScore($ip_address)
    {
        HookNode::where('ip_address', $ip_address)
            ->decrement('score', 1);
    }

    public static function incrementChunksProcessed($ip_address, $chunks_count = 1)
    {
        HookNode::where('ip_address', $ip_address)
            ->increment('chunks_processed_count', $chunks_count);
    }

    public static function setTimeOfLastContact($ip_address)
    {
        HookNode::where('ip_address', $ip_address)
            ->update('time_of_last_contact', Carbon::now());
    }
}
