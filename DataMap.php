<?php

include("Database.php");


/**
 * Builds and loads the datamap in DB given genesis hash and the number of file chunks
 *
 * @param string $genesis_hash A 64 char string
 * @param integer $file_chunk_count the number to generate
 * @return boolean $result response code
 */
function buildMap($genesis_hash, $file_chunk_count){
    
    
    $Db = initializeDbConnection();
    
    //we put the first chunk
    $chunk_id = 0;
    $Db->insertDataMapSection([[$genesis_hash, $chunk_id]]);
    
    //it takes the ceiling then disregards any past the correct number
    $hashes_per_db_update = 5;
    $num_groups = ceil($file_chunk_count / $hashes_per_db_update);
    
    //start process of getting hashes
    $hash_generator = hashGenerator($genesis_hash,  $file_chunk_count);
    
    for($i = 0; $i < $num_groups; $i++){
        
        $next_group = getNextNHashes($hash_generator, $hashes_per_db_update, $chunk_id);
        
        $chunk_id += $hashes_per_db_update;
        
        $Db->insertDataMapSection($next_group,$genesis_hash);
        
    }
    
    
}

/**
 * Separate out the Db initialization, this stuff will come from a config file
 */
function initializeDbConnection(){
    $server_name = 'localhost';
    $user_name = 'admin';
    $pass = 'nodeAdmin';
    $db_name = 'BrokerNodeDatabase';
    
    $DB = new BrokerDatabase($server_name, $user_name, $pass, $db_name);
    return $DB;
}

/**
 * Get the next group of hashes in the data map
 *
 * @param generator $hash_generator yields the next hash value in the sequence
 * @param integer $n the number to generate
 * @param $chunk_id the chunk id to start at
 * @return array Returns array of [chunk id , hash] pairs
 */
function getNextNHashes($hash_generator, $n, $chunk_id){
    
    $next_group = array();
    
    for($i = 0; $i < $n; $i++){
        
        $hash_generator->next();
        
        $next_hash = $hash_generator->current();
        
        //this is returned if the generator is called too many times
        if($next_hash == -1){
            break;
        }
        
        $next_hash = [$hash_generator->current(), $chunk_id++];
	
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
function hashGenerator($genesis_hash, $n){
    
    $hash = $genesis_hash;
    
    for($i = 0; $i < $n; $i++){
        
        yield $hash;
        
        $hash = hash("sha256", $hash);
        
    }
    while(True){
        yield -1;
    }
}




//informal test 
$genesishash = "sdfsdfsdfs";
buildMap($genesishash,  5);

?>
