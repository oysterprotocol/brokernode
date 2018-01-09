<?php

namespace App;

use Illuminate\Database\Eloquent\Model;

class UploadSession extends Model
{
    protected $table = 'upload_sessions';
    public $incrementing = false; // UUID
    protected $fillable = ['genesis_hash', 'file_size_bytes'];

}
