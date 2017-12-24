<?php
include("oy_config.php");

function oy_error($oy_error_code) {
    die($oy_error_code);
}

if (!isset($_POST['oy_hook_address'])) oy_error(1);

if (oy_db("SELECT `oy_hook_id` FROM `oy_hook` WHERE `oy_hook_status` = 1")->num_rows>=$oy_hook_limit) oy_error(2);

//validate the node

//if valid, add to hook pool