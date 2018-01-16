<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateDataMapTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('data_map', function (Blueprint $table) {
            $table->uuid('id');

            $table->string('genesis_hash', 255);
            $table->integer('chunk_idx')->unsigned();
            $table->string('hash', 255);

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
        Schema::dropIfExists('data_mappings');
    }
}
