<?php
namespace Controller;

require_once("Dao/BrokerNodeDAO.php");

use Dao\BrokerNodeDAO;

class BrokerNode {

	private $brokerNodeDao;

	function __construct() {
		$this->brokerNodeDao = new BrokerNodeDAO();
	}
	
	public function selectHookNode() {
        return $this->brokerNodeDao->getHigherScoreHookNode();
    }

	/**
	* Function to increase node's score
	* @author Arthur Mastropietro
	*/
	
	public function increaseHookNodeScore($nodeId) {
		return $this->brokerNodeDao->increaseHookNodeScore($nodeId);
	}
	
	/**
	* Function to decrease node's score
	* @author Arthur Mastropietro
	*/
	
	public function decreaseHookNodeScore($nodeId) {
		return $this->brokerNodeDao->decreaseHookNodeScore($nodeId);
	}
	
	/**
	* Function to change node's status
	* @author Arthur Mastropietro
	*/
	
	public function changeHookNodeStatus($nodeId, $status) {
		return $this->brokerNodeDao->changeHookNodeStatus($nodeId, $status);
	}
	
}

