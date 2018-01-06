<?php
require_once("IriWrapper.php");
require_once("../fromGit/brokernode/Iota Php port/PrepareTransfers.php");

$nl = "<br/>";
$nodeUrl = //PM REBEL FOR THIS
$apiVersion = '1.4';
$useFakeAddress = false;

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
$oysterSeed = 'OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL';

function checkAddress($address, $nodeUrl, $apiVersion) {
    $req = new IriWrapper($nodeUrl, $apiVersion);

    $command = new stdClass();
    $command->command = "findTransactions";
    $command->addresses = array($address);

    return $req->makeRequest($command);
}

echo ($nl . "findTransactions results" . $nl);
echo ($nl . "Using a ");
echo ($useFakeAddress ? "fake" : "real");
echo (" address:" . $nl);

$address = $useFakeAddress ? $fakeAddress : $realAddress;
$result = checkAddress($address, $nodeUrl, $apiVersion);
$txHasBeenConfirmed = count($result->hashes) != 0;

echo ($txHasBeenConfirmed ? "TX confirmed!  Do something else." : "Tx not confirmed!  Do POW!");

if (!$txHasBeenConfirmed) {

    $transactionObject = new stdClass();

    $transactionObject->address = $address;
    $transactionObject->value = 0;
    $transactionObject->message = $fakeData;



    //HAVE NOT DONE ANYTHING WITH THIS YET 

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


