<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateChunkEventsTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('chunk_events', function (Blueprint $table) {
            $table->uuid('id');
            $table->timestamps();
            $table->string('hooknode_id');
            $table->string('session_id');
            $table->string('event_name');
            $table->string('value');

            // Indexes
            $table->primary('id');
            $table->unique(['id']);
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('chunk_events');
    }
}

