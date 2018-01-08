<?php
require_once("IriWrapper.php");
require_once("PrepareTransfers.php");
require_once("IriData.php");
require_once("AttachmentCheck.php");
require_once("AttachTransaction.php");
require_once("HookNode.php");
require_once("BrokerNode.php");


$nodeUrl = // need to input this;
$useFakeAddress = true;

//Real
$realAddress = 'XSV99HBPZUXYAUABFOQYKJHNMHOSBCBAXFPNEQKWYLKQJMRFGGWRVIVSDGVVPSPGIQZMLBBSECC9PCZWN';

//Fakes
$fakeAddress = 'XSV99HBPZUXYAUABFOQYKJHNMHOSBCBAXFPNEQKWYLKQJMRFGGWRVIVSDGVVPSPGIQZMLBBSECC9PCZZZ';


$fakeData = 'SOMEFAKEDATATHATWEWANTTOPASSINFOROURTESTSBOYISUREHOPETHISWORKS';


$node = new BrokerNode($nodeUrl);

$transactionObject = new stdClass();

$transactionObject->address = $fakeAddress;
$transactionObject->message = 'SOMEFAKEDATATHATWEWANTTOPASSINFOROURTESTSBOYISUREHOPETHISWORKS';

$node->processNewData($transactionObject);
