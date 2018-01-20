<?php
namespace Controller;

require_once("Dao/HookNodeDAO.php");
require_once("Dao/BrokerNodeDAO.php");

use Dao\HookNodeDAO;
use Dao\BrokerNodeDAO;

class HookNode {

	private $hookNodeDao;
	private $brokerNodeDao;

	function __construct() {
		$this->hookNodeDao = new HookNodeDAO();
		$this->brokerNodeDao = new BrokerNodeDAO();
	}
	
	public function selectHookNode() {
        return $this->hookNodeDao->getHigherScoreNode();
    }

	/**
	* Function to increase node's score
	* @author Arthur Mastropietro
	*/
	
	public function increaseHookNodeScore($nodeId) {
		return $this->hookNodeDao->increaseNodeScore($nodeId);
	}
	
	/**
	* Function to decrease node's score
	* @author Arthur Mastropietro
	*/
	
	public function decreaseHookNodeScore($nodeId) {
		return $this->hookNodeDao->decreaseNodeScore($nodeId);
	}
	
	/**
	* Function to change node's status
	* @author Arthur Mastropietro
	*/
	
	public function changeHookNodeStatus($nodeId, $status) {
		return $this->hookNodeDao->changeHookNodeStatus($nodeId, $status);
	}
	
	
	/**
	* Function to verify with a given IpAddress exists
	* @author Arthur Mastropietro
	*/
	
	public function verifyRegisteredBrokerNode($ipAddress) {
		return $this->brokerNodeDao->verifyRegisteredBrokerNode($ipAddress);
	}
	
}

