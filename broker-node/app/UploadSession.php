<?php

namespace App;

use Illuminate\Database\Eloquent\Model;
use Webpatser\Uuid\Uuid;

class UploadSession extends Model
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

    protected $table = 'upload_sessions';
    public $incrementing = false; // UUID
    protected $fillable = [
        'genesis_hash',
        'file_size_bytes',
        'prl_payment',
        'payment_address'
    ];

}
