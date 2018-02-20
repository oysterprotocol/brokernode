<?php

namespace App;

use App\Clients\BrokerNode;
use Illuminate\Database\Eloquent\Model;
use Webpatser\Uuid\Uuid;

class Tips extends Model
{
    public static function boot()
    {
        parent::boot();
        self::creating(function ($model) {
            $model->id = (string)Uuid::generate(4);
        });
    }

    protected $table = 'tips';
    protected $fillable = [
        'tip',
        'node_id'
    ];

    public $incrementing = false;  // UUID

    public static function addTips($tips, $node_id = 'self')
    {
        if (!is_array($tips)) {
            $tips = array($tips);
        }

        foreach ($tips as $tip) {
            self::updateOrCreate(['tip' => $tip], ['node_id' => $node_id]);
        }
    }

    public static function getNextTips()
    {
        $tips = DB::table('tips')
            ->pluck('tip')
            ->take(2);

        echo 'getting next tips';
        var_dump($tips);

        DB::table('tips')
            ->whereIn('tip', $tips)
            ->delete();
    }
}
