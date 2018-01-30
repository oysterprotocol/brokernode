<?php

namespace App;

use Illuminate\Database\Eloquent\Model;
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
            $model->id = (string) Uuid::generate(4);
        });
    }

    protected $table = 'hook_nodes';
    protected $fillable = ['ip_address'];
    public $incrementing = false; // UUID

    public static function insertNode($ip_address) {
        return self::create(['ip_address' => $ip_address]);
    }

    public static function getNextReadyNode() {
        self::where('status', "ready")
            ->orderBy('score', 'desc')
            ->first();
    }

    public static function incrementScore($ip_address) {
        self::where('ip_address', $ip_address)
            ->increment('score', 1);
    }

    public static function decrementScore($ip_address) {
        self::where('ip_address', $ip_address)
            ->decrement('score', 1);
    }

    public static function incrementChunksProcessed($ip_address, $chunks_count=1) {
        self::where('ip_address', $ip_address)
            ->increment('chunks_processed_count', $chunks_count);
    }
}
