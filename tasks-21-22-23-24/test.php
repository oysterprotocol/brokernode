<?php 

require_once("Controller/HookNode.php");

$hookNode = new Controller\HookNode();

echo "Test: Check hook node with higher score:<br/>";
$node = $hookNode->selectHookNode();
echo "<pre>";
print_r($node);
echo "</pre>";
echo "<br/><br/>";

echo "Test: Increase hook node score:<br/>";
$hookNode->increaseHookNodeScore(1);

echo "Test: Decrease hook node score:<br/>";
$hookNode->decreaseHookNodeScore(2);

echo "Test: Change hook node status:<br/>";
$hookNode->changeHookNodeStatus(1, 2);

echo "Test: Check if a broker node exists:<br/>";
$ok = $hookNode->verifyRegisteredBrokerNode("164.247.37.148");
echo $ok;
echo "<br/><br/>";
