<?php

namespace App;

use Illuminate\Database\Eloquent\Model;

class DataMapping extends Model
{
    protected $table = 'data_mappings';
}

// this is a MariaDB based data map loading module
// IN PROGRESS

//make into generator  *sunday
function hashGenerator($genesis_hash,$n){

	$hash = $genesis_hash;

	for($i = 0; $i < $n; $i++){

		yield $hash;

		$hash = hash("sha256", $hash);

	}
}


function getNextNHashes($hash_generator, $n, $chunk_id){

    $next_group = array();

    for($i = 0; $i < $n; $i++){

        $hash_generator->next();
        $next_hash = [$hash_generator->current(), $chunk_id++];

        array_push($next_group, $next_hash);

    }

    return $next_group;
}


//test
//chunk count is file chunk count

function buildMap($genesis_hash, $file_chunk_count){

	//the first call to hashgenerator returns the genesis_hash as we put it in db first
	$chunk_id = 0;
	$next_hash = "";
    $hashes_per_db_update = 5;
    $num_groups = ceil($file_chunk_count / $hashes_per_db_update);

	//start process of getting hashes
	$hash_generator = hashGenerator($genesis_hash,  $file_chunk_count);

	for($i = 0; $i < $num_groups; $i++){

	    $next_group = getNextNHashes($hash_generator, $hashes_per_db_update, $chunk_id);

	    $chunk_id += $hashes_per_db_update;

	    print("next group");
	    var_dump($next_group);

	    //TODO insert into DB

	}


}

//informal test

$genesishash = "sdfsdfsdfs";

buildMap($genesishash,  5);
