<?php
require_once("../Dao/BrokerNodeDAO.php");


$dao = new Dao\BrokerNodeDAO();
for($i = 0; $i < 99; $i++){
	$ip = "".mt_rand(0,255).".".mt_rand(0,255).".".mt_rand(0,255).".".mt_rand(0,255);
	
	$dao->insertNode(array("ipAddress" => $ip));
}