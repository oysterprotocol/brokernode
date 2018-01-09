<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateDataMappingsTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('data_mappings', function (Blueprint $table) {
            $table->increments('id');

            $table->string('genesis_hash', 255);
            $table->string('hash', 255);
            $table->integer('chunk_idx')->unsigned();

            $table->timestamps();

            // Indexes
            $table->unique('genesis_hash');
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
