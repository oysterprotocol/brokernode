<?php
namespace Model;

class HookNodeModel {

	private $ipAddress;
	private $timestamp;
	private $score;
	private $status;
	
	public function setIpAddress($__ipAddress){
		$this->ipAddress = $__ipAddress;
	}
	
	public function setTimestamp($__timestamp){
		$this->timestamp = $__timestamp;
	}
	
	public function setScore($__score){
		$this->score = $__score;
	}
	
	public function setStatus($__status){
		$this->status = $__status;
	}
	
	public function getIpAddress() {
		return $this->ipAddress;
	}
	
	public function getTimestamp() {
		return $this->timestamp;
	}
	
	public function getScore() {
		return $this->score;
	}
	
	public function getStatus() {
		return $this->status;
	}
}