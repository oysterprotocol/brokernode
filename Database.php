<?php 

class BrokerDatabase {
    
    
    public function openDatabaseConnection(){
        $server_name = 'localhost';
        $user_name = 'root';
        $pass = 'nodeAdmin';
        $db_name = "BrokerNodeDatabase";
        $connection = mysqli_connect($server_name, $user_name, $pass, $db_name);
        if ($connection->connect_error) {
            die("Connection failed: " . $conn->connect_error);
        }
        return $connection;
    }
    
    public function closeDatabaseConnection($connection){
        $connection->close();
    }
    
    public function processSql($connection, $sqlString){
        return $connection->query($sqlString);
    }
    
    //SESSION MANAGEMENT
    function initializeSessionInDatabase($session_id, $genesis_hash){
        
        //TODO
    }
    
    //DATA MAP INSERTION---------------------------------------------------
    public function insertDataMapSection($hashes, $genesis_hash){
        
        print("enteringDM");
        $connection = $this->openDatabaseConnection();
        
        $sqlString = $this->buildDataMapInsertStatement($hashes, $genesis_hash);
       
        $result = $this->processSql($connection, $sqlString);
             
        $this->closeDatabaseConnection($connection);
        
        return $result;
        
    }
    
    //fix last sql prob tomorrow
    public function buildDataMapInsertStatement($hash_mappings, $genesis_hash){
        
        $sqlString = "INSERT INTO DataMapping (genesis_hash, chunk_id, hash) VALUES ";
        
        $sqlChunks = array_map(
            
            //array($this,"buildChunkInsertSnippet"),
            
            function($hash_mapping) use ($genesis_hash){
                
                print("gen inside");
                print($genesis_hash);
                $snippet = "(" . $genesis_hash . "," . strval($hash_mapping[0]) . "," . strval($hash_mapping[1]) . ")";
                print($snippet);
                return $snippet;
            },
            
            $hash_mappings);
        
        print("before explode");
        print($sqlString);
        $sqlString = $sqlString . "".implode($sqlChunks) . ";";
        print("afterexplode");
        print($sqlString);
        
        return $sqlString;
    }
    
   
        
}

//TEST

function testDataMapInsert(){
    
    print("beginning test\n");
    $DB = new BrokerDatabase();
    
    $hashes  = [["f23", 1, "ddcsdc"], ["f3f3", 2, "dsfsdf"]];
    
    $result = $DB->insertDataMapSection($hashes, "teqwer");
    
    print("result\n");
    print($result);
}

testDataMapInsert();

?>





