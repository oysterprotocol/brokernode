<?php

namespace App;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Facades\DB;
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
        if (DB::table('tips')->count() < 2) {
            return null;
        }

        $tipObjs = DB::table('tips')
            ->orderBy('node_id')
            ->take(2)
            ->get()
            ->toArray();

        $tips = array($tipObjs[0]->tip, $tipObjs[1]->tip);

        DB::table('tips')
            ->whereIn('tip', $tips)
            ->delete();

        if ($tipObjs[0]->node_id != $tipObjs[0]->node_id) {
            return self::getNextTips();
        } else {
            return $tips;
        }
    }
}
