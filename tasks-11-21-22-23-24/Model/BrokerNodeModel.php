<?php
namespace Model;

class BrokerNodeModel {
	
	private $ipAddress;
	
	public function setIpAddress($__ipAddress){
		$this->ipAddress = $__ipAddress;
	}
	
	public function getIpAddress() {
		return $this->ipAddress;
	}
	
}