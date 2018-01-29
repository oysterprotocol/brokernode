<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateUploadSessionsTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('upload_sessions', function (Blueprint $table) {
            $table->uuid('id');


            // references genesis_hash on data_mappings
            $table->string('genesis_hash', 255)
                  ->unique();
            $table->unsignedBigInteger('file_size_bytes');
            $table->enum('type', [
                    'alpha',
                    'beta'
                ])
                ->default('alpha');

            $table->timestamps();

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
        Schema::dropIfExists('upload_sessions');
    }
}
