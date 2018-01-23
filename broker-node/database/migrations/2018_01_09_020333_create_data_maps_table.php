<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateDataMapsTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('data_maps', function (Blueprint $table) {
            $table->uuid('id');

            $table->string('genesis_hash', 255);
            $table->integer('chunk_idx')->unsigned();
            $table->string('hash', 255);
            $table->binary('chunk');
            $table->string('hooknode_id');
            $table->enum('status', [
                    'unassigned',
                    'pending',
                    'complete',
                    'error'
                ])
                ->default('unassigned');  // TODO: Use integer mapping.

            $table->timestamps();

            // Indexes
            $table->primary('id');
            $table->unique(['genesis_hash', 'chunk_idx']);
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('data_maps');
    }
}
