<?php
namespace Controller;

require_once("Dao/hookNodeDao.php");

use Dao\HookNodeModel;

class BrokerNode {

	private $hookNodeDao;

	function __construct() {
		$this->hookNodeDao = new hookNodeDao();
	}
	
	public function selectHookNode() {
        return $this->hookNodeDao->getHigherScoreHookNode();
    }

	/**
	* Function to increase node's score
	* @author Arthur Mastropietro
	*/
	
	public function increaseHookNodeScore($nodeId) {
		return $this->hookNodeDao->increaseHookNodeScore($nodeId);
	}
	
	/**
	* Function to decrease node's score
	* @author Arthur Mastropietro
	*/
	
	public function decreaseHookNodeScore($nodeId) {
		return $this->hookNodeDao->decreaseHookNodeScore($nodeId);
	}
	
	/**
	* Function to change node's status
	* @author Arthur Mastropietro
	*/
	
	public function changeHookNodeStatus($nodeId, $status) {
		return $this->hookNodeDao->changeHookNodeStatus($nodeId, $status);
	}
	
	/**
	* Function to increase num of num chunks processed by hookNode
	* @params $id: hookNode Id, $num: number to be increased
	* @author Arthur Mastropietro
	*/
	public function increaseNumChunksProcessed($id, $num) {
		return $this->hookNodeDao->increaseNumChunksProcessed($id, $num);
	}
}

