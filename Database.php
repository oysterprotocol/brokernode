<?php 

/**
 * Broker Database
 *
 * This class gives access to the broker database.  There are three tables:
 *  upload session, data mappings, and hook node directory.
 *  The class methods pertain to starting and stopping sessions, 
 *  building and deleting data maps, and looking up hook nodes.
 *
 * @author     Oyster 
 */
class BrokerDatabase {
        
    /**
     * Constructor
     */
    function __construct($server_name,$user_name,$pass,$db_name) {
        
        $this->server_name = $server_name;
        $this->user_name = $user_name;
        $this->pass = $pass;
        $this->db_name = $db_name;
        
    }
    
    
    /**
     * Open/CLose
     */
    public function openDatabaseConnection(){
        
        $connection = mysqli_connect($server_name, $user_name, $pass, $db_name);
        if ($connection->connect_error) {
            die("Connection failed: " . $conn->connect_error);
        }
        return $connection;
    }
    
    public function closeDatabaseConnection($connection){
        $connection->close();
    }
    
    /**
     * SESSION MANAGEMENT
     */
    function initializeSessionInDatabase($session_id, $genesis_hash){        
        //TODO
    }
    
    /**
     * Inserts an array of hash value to chunk id pairs along with the genesis hash.
     * @param array $hash_chunkid_pairs [['asdasd',1],...]
     * @return int indicating success or failues
     * @throws Exception If element in array is not an integer
     */
    public function insertDataMapSection($hash_chunkid_pairs, $genesis_hash){
        
        print("enteringDM");
        
        try {
            
            $connection = $this->openDatabaseConnection();
            
            $sqlString = $this->buildDataMapInsertStatement($hash_chunkid_pairs, $genesis_hash);
            
            $result = $connection->query($sqlString);
            
        } catch (Exception $e) {
           
            echo $e->getMessage();
        }

        $this->closeDatabaseConnection($connection);
        
        return $result;
    }
    

    public function buildDataMapInsertStatement($hash_mappings, $genesis_hash){
        
        $sqlString = "INSERT INTO DataMapping (genesis_hash, hash, chunk_id) VALUES ";
        
        
        $sqlChunks = array_map(
            
            //array($this,"buildChunkInsertSnippet"),
            
            function($hash_mapping) use ($genesis_hash){
                
                //given x1,x2,x3 this function creates a string '("x1","x2","x3")'
                $snippet = "(\"" . $genesis_hash . "\",\"" . strval($hash_mapping[0]) . "\",\"" . strval($hash_mapping[1]) . "\"),";
            },
            
            $hash_mappings);
        
        //put this pieces together
        $sqlString = $sqlString . "".implode($sqlChunks);
        
        //trim last comma
        $sqlString = rtrim($sqlString,',');
        
        //and finish
        $sqlString = $sqlString . ";";
        
        return $sqlString;
    }    
}

//TEST
function testDataMapInsert(){
    
    
    $server_name = 'localhost';
    $user_name = 'admin';
    $pass = 'nodeAdmin';
    $db_name = 'BrokerNodeDatabase';
    
    print("beginning test\n");
    $DB = new BrokerDatabase($server_name, $user_name, $pass, $db_name);
    
    $hashes  = [["ea", 1], ["we", 2]];
    
    $result = $DB->insertDataMapSection($hashes, "asd");
    
    print("result\n");
    print($result);
}

testDataMapInsert();

?>





