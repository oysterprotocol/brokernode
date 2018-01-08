<?php

require_once("IriWrapper.php");
require_once("IriData.php");
require_once("PrepareTransfers.php");

class AttachTransaction
{
    public $txTrytes = '';

    public function __construct()
    {
    }

    public function attachTx($dataObject)
    {
        $trytesToBroadcast = NULL;

        $transactionObject = new stdClass();

        $transactionObject->address = $dataObject->address;
        $transactionObject->value = $GLOBALS['txValue'];
        $transactionObject->message = $dataObject->message;
        $transactionObject->tag = $GLOBALS['nodeId'];

        try {
            $transactionObject->trytes = PrepareTransfers::buildTxTrytes($dataObject);
        } catch (Exception $e) {
            echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
        }

        if (!is_null($transactionObject->trytes)) {
            try {
                self::getTransactionsToApprove($transactionObject);
            } catch (Exception $e) {
                echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
            }
        }

        if (property_exists($transactionObject, 'trunkTransaction')) {
            try {
                $trytesToBroadcast = self::attachToTangle($transactionObject);
            } catch (Exception $e) {
                echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
            }
        }

        if (!is_null($trytesToBroadcast)) {
            try {
                self::broadcastTransactions($trytesToBroadcast);
            } catch (Exception $e) {
                echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
            }
        }
    }

    private function getTransactionsToApprove(&$transactionObject)
    {
        $req = new IriWrapper();

        $command = new stdClass();
        $command->command = "getTransactionsToApprove";
        $command->depth = $GLOBALS['depthToSearchForTxs'];

        $result = $req->makeRequest($command);

        if (!is_null($result)) {
            //switching trunk and branch
            $transactionObject->trunkTransaction = $result->branchTransaction;
            $transactionObject->branchTransaction = $result->trunkTransaction;
        }
        else {
            throw new Exception('getTransactionToApprove failed!');
        }
    }

    private function attachToTangle($transactionObject)
    {
        $req = new IriWrapper();

        $command = new stdClass();
        $command->command = "attachToTangle";
        $command->minWeightMagnitude = $GLOBALS['minWeightMagnitude'];
        $command->trunkTransaction = $transactionObject->trunkTransaction;
        $command->branchTransaction = $transactionObject->branchTransaction;
        $command->trytes = $transactionObject->trytes;

        $resultOfAttach = $req->makeRequest($command);

        if (!is_null($resultOfAttach) && property_exists($resultOfAttach, 'trytes')) {
            return $resultOfAttach->trytes;
        } else {
            throw new Exception('attachToTangle failed!');
        }
    }

    private function broadcastTransactions($trytesToBroadcast) {
        $req = new IriWrapper();

        $command = new stdClass();
        $command->command = "broadcastTransactions";
        $command->trytes = $trytesToBroadcast;

        $resultOfBroadcast = $req->makeRequest($command);

        var_dump($resultOfBroadcast);

//        if (property_exists($resultOfAttach, 'trytes')) {
//            $this->broadcastTransactions($resultOfAttach);
//        } else {
//            throw new Exception('attachToTangle failed!');
//        }
    }
}