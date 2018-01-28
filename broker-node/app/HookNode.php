<?php

namespace App;

use Illuminate\Database\Eloquent\Model;

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
}
