<?php
require_once("IriWrapper.php");
require_once("PrepareTransfers.php");

$nl = "<br/>";
$nodeUrl = 'http://104.5.45.19:14265';
$apiVersion = '1.4';
$useFakeAddress = true;

//Real
$realAddress = 'XSV99HBPZUXYAUABFOQYKJHNMHOSBCBAXFPNEQKWYLKQJMRFGGWRVIVSDGVVPSPGIQZMLBBSECC9PCZWN';
$realBundleHash = 'K9BS9GPYA9IQYHKQRKVDCDDIUISUYTY9BXUNAXLRDSJJNRK9FJFFVLIKARLOFZSKRCERRLUOHQSAGOELD';
$realTxHash = 'BYMFJLUEASMHBIXAYUXFOJMVPCPSYJPYNTJ9JCKLCCZLCYDKUROTOLMIRP9ZOLXGRCOKHTOUGRHSA9999';
//Fakes
$fakeAddress = 'XSV99HBPZUXYAUABFOQYKJHNMHOSBCBAXFPNEQKWYLKQJMRFGGWRVIVSDGVVPSPGIQZMLBBSECC9PCZZZ';
$fakeBundleHash = 'K9BS9GPYA9IQYHKQRKVDCDDIUISUYTY9BXUNAXLRDSJJNRK9FJFFVLIKARLOFZSKRCERRLUOHQSAGOZZZ';
$fakeTxHash = 'BYMFJLUEASMHBIXAYUXFOJMVPCPSYJPYNTJ9JCKLCCZLCYDKUROTOLMIRP9ZOLXGRCOKHTOUGRHSA9ZZZ';

$fakeData = 'SOMEFAKEDATATHATWEWANTTOPASSINFOROURTESTSBOYISUREHOPETHISWORKS';
$prepareTransfersResult;


echo ($nl . "findTransactions results" . $nl);
echo ($nl . "Using a ");
echo ($useFakeAddress ? "fake" : "real");
echo (" address:" . $nl);

$address = $useFakeAddress ? $fakeAddress : $realAddress;
$result = NULL;

while ($result == NULL) {
    $result = checkAddress($address, $nodeUrl, $apiVersion);
}

var_dump($result);

$txHasBeenConfirmed = !(is_null($result->hashes) || count($result->hashes) == 0);

echo $nl . ($txHasBeenConfirmed ? "TX confirmed!  Do something else." : "Tx not confirmed!  Do POW!") . $nl;

if (!$txHasBeenConfirmed) {

    $transactionObject = new stdClass();

    $transactionObject->address = $address;
    $transactionObject->value = 0;
    $transactionObject->message = $fakeData;
    $transactionObject->tag = 'OYSTERNODE';

    PrepareTransfers::prepareTransfers($oysterSeed,
        [$transactionObject],
        null, //where options with inputs array would go, consider removing this and removing param from method
        function ($e, $s) {
            if ($s != null) {
                $prepareTransfersResult = $s;
                //do something with $s, which should be an array of transaction trytes
            } else {
                echo $e;
                //do something with this error
            }
        });
}







nodeSetup($nodeUrl, $apiVersion);

echo($nl . "findTransactions results" . $nl);
echo($nl . "Using a ");
echo($useFakeAddress ? "fake" : "real");
echo(" address:" . $nl);

$address = $useFakeAddress ? $fakeAddress : $realAddress;
$result = NULL;

while ($result == NULL) {
    $result = checkAddress($address, $nodeUrl, $apiVersion);
}

var_dump($result);

$txHasBeenConfirmed = !(is_null($result->hashes) || count($result->hashes) == 0);

echo $nl . ($txHasBeenConfirmed ? "TX confirmed!  Do something else." : "Tx not confirmed!  Do POW!") . $nl;

if (!$txHasBeenConfirmed) {

    $transactionObject = new stdClass();

    $transactionObject->address = $address;
    $transactionObject->value = 0;
    $transactionObject->message = $fakeData;
    $transactionObject->tag = 'OYSTERNODE';

    PrepareTransfers::prepareTransfers($oysterSeed,
        [$transactionObject],
        null, //where options with inputs array would go, consider removing this and removing param from method
        function ($e, $s) {
            if ($s != null) {
                $prepareTransfersResult = $s;
                //do something with $s, which should be an array of transaction trytes
            } else {
                echo $e;
                //do something with this error
            }
        });
}

