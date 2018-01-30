<?php
namespace Dao;

require_once("Lib/Database.php");
require_once("Model/HookNodeModel.php");

use Lib\Database;
use Model\HookNodeModel;

class HookNodeDAO extends Database {
	
	function __construct() {
		parent::__construct();
	}
	
	public function insertNode($data) {
		$this->bindMore(array("ip_address" => $data["ipAddress"]));
		$insert = $this->query("INSERT INTO brokernode(ip_address) VALUES(:ip_address)");
		return $insert;
	}
	
	public function getHigherScoreHookNode() {
		$node = $this->query("SELECT id, ip_address, timestamp, score from hooknode ORDER BY score DESC, timestamp DESC LIMIT 1");
		$model = new HookNodeModel();
		$singleNode = $node[0];
		$model->setIpAddress($singleNode["ip_address"]);
		$model->setTimestamp($singleNode["timestamp"]);
		$model->setScore($singleNode["score"]);
		return $model;
	}
	
	public function increaseHookNodeScore($id) {
		$update = $this->query("UPDATE hooknode SET score = score + 1 WHERE id = :id", array("id" => $id));
		return $update;
	}
	
	public function decreaseHookNodeScore($id) {
		$update = $this->query("UPDATE hooknode SET score = score - 1 WHERE id = :id", array("id" => $id));
		return $update;
	}
	
	public function insertHookNode($data) {
		$this->bindMore(array("ip_address" => $data["ipAddress"], "timestamp" => $data["timestamp"], "score" => $data["score"]));
		$insert = $this->query("INSERT INTO hooknode(ip_address, timestamp, score) VALUES(:ip_address, :timestamp, :score)");
		return $insert;
	}
	
	public function increaseNumChunksProcessed($id, $num) {
		return $this->hookNodeDao->increaseNumChunksProcessed($num);
		$update = $this->query("UPDATE hooknode SET num_chunks_processed = num_chunks_processed + :num WHERE id = :id", array("id" => $id, 'num' => $num));
		return $update;
	}
	
	function __destruct() {
		$this->closeConnection();
	}
	
}