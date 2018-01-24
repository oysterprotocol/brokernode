<?php
require_once("../Dao/HookNodeDAO.php");


$dao = new HookNodeDAO();
for($i = 0; $i < 99; $i++){
	$ip = "".mt_rand(0,255).".".mt_rand(0,255).".".mt_rand(0,255).".".mt_rand(0,255);
	$score = rand(0,1100);
	$timestamp = date('Y-m-d', strtotime( '+'.mt_rand(0,30).' days'));
	
	$dao->insertNode(array("ipAddress" => $ip, "timestamp" => $timestamp, "score" => $score));
}