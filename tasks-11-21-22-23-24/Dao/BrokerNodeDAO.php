<?php
namespace Dao;

require_once("Lib/Database.php");
require_once("Model/BrokerNodeModel.php");

use Lib\Database;
use Model\BrokerNodeModel;

class BrokerNodeDAO extends Database {
	
	function __construct() {
		parent::__construct();
	}
	
	public function insertNode($data) {
		$this->bindMore(array("ip_address" => $data["ipAddress"]));
		$insert = $this->query("INSERT INTO brokernode(ip_address) VALUES(:ip_address)");
		return $insert;
	}
	
	public function verifyRegisteredBrokerNode($ipAddress) {
		$node = $this->query("SELECT id from brokernode WHERE ip_address = :ip_address", array("ip_address" => $ipAddress));
		if (count($node) > 0){
			return true;
		}
		return false;
	}
	
	function __destruct() {
		$this->closeConnection();
	}
}