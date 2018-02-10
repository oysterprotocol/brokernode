<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateHookNodesTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('hook_nodes', function (Blueprint $table) {
            $table->uuid('id');
            $table->timestamps();

            $table->string('ip_address', 255)
                ->unique();
            $table->unsignedBigInteger('score')
                ->default(0);
            $table->unsignedBigInteger('chunks_processed_count')
                ->default(0);
            $table->unsignedBigInteger('time_of_last_chunk')
                ->default(0);
            $table->enum('status', [
                'ready',
                'processing',
            ])
                ->default('ready');

            // Indexes
            $table->primary('id');
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('hook_nodes');
    }
}
