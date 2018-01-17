<?php
include("oy_config.php");

$oy_iri_port = "14500";

function oy_error($oy_error_code) {
    die($oy_error_code);
}

function oy_hook_invoke($oy_hook_address, $oy_command) {
    return file_get_contents("http://".$oy_hook_address."/?oy_command=".$oy_command);
}

function oy_iri($oy_url, $oy_command) {
    $oy_headers = array(
        "Content-Type:application/json",
        "X-IOTA-API-Version: 1.4"
    );

    $ch = curl_init();

    curl_setopt($ch, CURLOPT_HEADER, 0);
    curl_setopt($ch, CURLOPT_HTTPHEADER, $oy_headers);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
    curl_setopt($ch, CURLOPT_URL, $oy_url);
    curl_setopt($ch, CURLOPT_POST, 1);
    curl_setopt($ch, CURLOPT_POSTFIELDS, "{\"command\":\"".$oy_command."\"}");

    $data = curl_exec($ch);
    curl_close($ch);

    return $data;
}

if (!isset($_POST['oy_hook_address'])) oy_error(1);

//if (oy_db("SELECT `oy_hook_id` FROM `oy_hook` WHERE `oy_hook_status` = 1")->num_rows>=$oy_hook_limit) oy_error(2);

$oy_command = "getNeighbors";

$oy_hook_info = oy_hook_invoke($_POST['oy_hook_address'], "getNodeInfo");

$oy_local_info = oy_iri("http://localhost:".$oy_iri_port, $_GET['oy_command']);

//TODO make sure the milestones defined in $oy_hook_info match with $oy_local_info.

//TODO if the hook node actively rejects the connection, then drop it from the waiting list. If it's simply out of sync, or responding yet not qualifying - then reset it's position in the pool


$oy_hook_payout = oy_hook_invoke($_POST['oy_hook_address'], "getNodePayout");

//TODO record the payout ETH address in the database, along with the node as a valid/active node