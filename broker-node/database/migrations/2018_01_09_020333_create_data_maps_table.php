<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;
use App\DataMap;

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

            $table->binary('chunk')->nullable();
            $table->string('hooknode_id')->nullable();
            $table->string('address');
            $table->string('message');
            $table->string('trunkTransaction');
            $table->string('branchTransaction');
            $table->enum('status', [
                DataMap::status['unassigned'],
                DataMap::status['pending'],
                DataMap::status['unverified'],
                DataMap::status['complete'],
                DataMap::status['error']
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
