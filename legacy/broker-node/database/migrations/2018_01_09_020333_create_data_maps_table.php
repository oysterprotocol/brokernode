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
            $table->enum('status', [
                DataMap::status['unassigned'],
                DataMap::status['unverified'],
                DataMap::status['complete'],
                DataMap::status['error']
            ])->default(DataMap::status['unassigned']);  // TODO: Use integer mapping.
            $table->string('hooknode_id')->nullable();
            $table->mediumText('message')->nullable();  // ~16mb limit.
            $table->string('trunkTransaction')->nullable();
            $table->string('branchTransaction')->nullable();
            $table->uuid('id');
            $table->string('genesis_hash', 255);
            $table->integer('chunk_idx')->unsigned();
            $table->string('hash', 255);
            $table->string('address')->nullable();

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
