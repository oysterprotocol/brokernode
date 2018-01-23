<?php

namespace App\Console\Commands;

use App\DataMap;
use Illuminate\Console\Command;

class CheckChunkStatus extends Command {
    protected $signature = 'CheckChunkStatus:checkStatus';
    protected $description =
        'Polls the status of chunks that have been sent to hook nodes';

    /**
     * Create a new command instance.
     */
    public function __construct() {
        parent::__construct();
    }

    /**
     * Execute the console command.
     */
    public function handle() {
        echo DataMap::all();
    }
}
