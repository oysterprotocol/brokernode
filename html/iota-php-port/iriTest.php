<?php
require_once("IriWrapper.php");
require_once("PrepareTransfers.php");
require_once("IriData.php");
require_once("HookNode.php");
require_once("ChunkProcess.php");

$useFakeAddress = true;

//Real
$realAddress = 'XSV99HBPZUXYAUABFOQYKJHNMHOSBCBAXFPNEQKWYLKQJMRFGGWRVIVSDGVVPSPGIQZMLBBSECC9PCZWN';

//Fakes
$fakeAddress = 'XSV99HRRZUXYAUABFOQYKJHNMHOSBCBAXFPNEEKWWEKQJMRFGGWRVIVSDGVVPSPGIQZMLBBSECC9PCZZZ';


$fakeData = 'SOMEFAKEDATATHATWEWANTTOPASSINFOROURTESTSBOYISUREHOPETHISWORKS';

$node = new ChunkProcess($nodeUrl);

$transactionObject = new stdClass();

$transactionObject->address = $fakeAddress;
$transactionObject->message = 'SOMEFAKEDATATHATWEWANTTOPASSINFOROURTESTSBOYISUREHOPETHISWORKS';

//$node->processNewData($transactionObject);



$iri = new IriWrapper();

$commandObj = new stdClass();

$headers = array(
    'Content-Type: application/json',
);

$userAgent = 'Codular Sample cURL Request';
$apiVersionHeaderString = 'X-IOTA-API-Version: ';

//$commandObj->command = 'getNeighbors';
//$commandObj->command = 'addNeighbors';
//$commandObj->command = 'getNodeInfo';
//$commandObj->command = 'getNodeInfo';
$commandObj->test = "WHOOOOOOHOOOOOO";

//$iri->makeRequest($commandObj);

    $payload = json_encode($commandObj);

    $curl = curl_init();

    curl_setopt_array($curl, array(
        //CURLOPT_RETURNTRANSFER => 1,
        CURLOPT_POST => 1,
        CURLOPT_URL => $nodeUrl,
        CURLOPT_USERAGENT => $userAgent,
        CURLOPT_POSTFIELDS => $payload,
        CURLOPT_HTTPHEADER => $headers,
        CURLOPT_CONNECTTIMEOUT => 0,
        CURLOPT_TIMEOUT => 1000
    ));

    echo $GLOBALS['nl'];
    echo $GLOBALS['nl'] . "calling curl, url is: " . $nodeUrl . $GLOBALS['nl'];
    //echo "command: " . $commandObject->command . $GLOBALS['nl'];
    //echo $GLOBALS['nl'] . "payload: " . $payload . $GLOBALS['nl'];

    $response = json_decode(curl_exec($curl));
    curl_close($curl);

    echo $GLOBALS['nl'] . "response was: " . $GLOBALS['nl'];
    var_dump($response);

    return $response;

