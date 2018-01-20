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
	
	public function getHigherScoreNode() {
		$node = $this->query("SELECT id, ip_address, timestamp, score from hooknode ORDER BY score DESC, timestamp DESC LIMIT 1");
		$model = new HookNodeModel();
		$singleNode = $node[0];
		$model->setIpAddress($singleNode["ip_address"]);
		$model->setTimestamp($singleNode["timestamp"]);
		$model->setScore($singleNode["score"]);
		return $model;
	}
	
	public function increaseNodeScore($id) {
		$update = $this->query("UPDATE hooknode SET score = score + 1 WHERE id = :id", array("id" => $id));
		return $update;
	}
	
	public function decreaseNodeScore($id) {
		$update = $this->query("UPDATE hooknode SET score = score - 1 WHERE id = :id", array("id" => $id));
		return $update;
	}
	
	public function insertNode($data) {
		$this->bindMore(array("ip_address" => $data["ipAddress"], "timestamp" => $data["timestamp"], "score" => $data["score"]));
		$insert = $this->query("INSERT INTO hooknode(ip_address, timestamp, score) VALUES(:ip_address, :timestamp, :score)");
		return $insert;
	}
	
	public function changeHookNodeStatus($id, $status) {
		$update = $this->query("UPDATE hooknode SET status = :status WHERE id = :id", array("id" => $id, 'status' => $status));
		return $update;
	}
	
	function __destruct() {
		$this->closeConnection();
	}
	
}