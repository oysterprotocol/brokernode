<?php

namespace App;

use Illuminate\Database\Eloquent\Model;

class DataMapping extends Model
{
	protected $table = 'data_mappings';
	protected $fillable = ['genesis_hash', 'hash', 'chunk_idx'];
}

/**
 * Builds and loads the datamap in DB given genesis hash and the number of file chunks
 *
 * @param string $genesis_hash A 64 char string
 * @param integer $file_chunk_count the number to generate
 * @return boolean $result response code
 */
function buildMap($genesis_hash, $file_chunk_count) {
    //we put the first chunk
    $chunk_idx = 0;
    App\DataMapping::create([
        'genesis_hash' => $genesis_hash,
        'hash' => $genesis_hash,
        'chunk_idx' => $chunk_idx
    ]);

    //it takes the ceiling then disregards any past the correct number
    $hashes_per_db_update = 5;
    $num_groups = ceil($file_chunk_count / $hashes_per_db_update);

    //start process of getting hashes
    $hash_generator = hashGenerator($genesis_hash,  $file_chunk_count);

    for($i = 0; $i < $num_groups; $i++) {
        $next_group = getNextNHashes($hash_generator, $hashes_per_db_update, $chunk_idx);
        $chunk_idx += $hashes_per_db_update;

        foreach($next_group as $chunk_idx_hash) {
            [$curr_chunk_idx, $curr_hash] = $chunk_idx_hash;

            App\DataMapping::create([
                'genesis_hash' => $genesis_hash,
                'hash' => $curr_hash,
                'chunk_idx' => $curr_chunk_idx
            ]);
        }
    }
}

/**
 * Get the next group of hashes in the data map
 *
 * @param generator $hash_generator yields the next hash value in the sequence
 * @param integer $n the number to generate
 * @param $chunk_idx the chunk id to start at
 * @return array Returns array of [chunk id , hash] pairs
 */
function getNextNHashes($hash_generator, $n, $chunk_idx) {
	$next_group = array();

    for($i = 0; $i < $n; $i++) {
        $hash_generator->next();
        $next_hash = $hash_generator->current();
        //this is returned if the generator is called too many times
        if($next_hash == -1) {
            break;
        }

        $next_hash = [$hash_generator->current(), $chunk_idx++];
        array_push($next_group, $next_hash);
    }

    return $next_group;
}

/**
 * A generator that yields the next of the data map by rehashing the last hash
 * If we call it more than $n times it will just return -1
 *
 * @param string $genesis_hash A 64 char string
 * @param integer $n the number to generate
 * @return User Returns User object or null if not found
 */
function hashGenerator($genesis_hash, $n) {
    $hash = $genesis_hash;

    for($i = 0; $i <= $n; $i++) {
        yield $hash;
        $hash = hash("sha256", $hash);

    }
    while(True) {
        yield -1;
    }
}
