<?php

namespace App\Clients;

require_once("iota-php-port/PrepareTransfers.php");
require_once("requests/IriData.php");
require_once("requests/IriWrapper.php");
require_once("requests/NodeMessenger.php");

// This is a temporary hack to make the above required files work in this
// namespace. We can clean this up after testnet.
use \Exception;
use \IriData;
use \IriWrapper;
use \NodeMessenger;
use \PrepareTransfers;
use \stdClass;

class BrokerNode
{
    public static $chunksToAttach = null;
    public static $chunksToVerify = null;

    public static $IriWrapper = null;
    public static $NodeMessenger = null;

    public static $iriRequestInProgress = false;


    private static function initIri()
    {
        if (is_null(self::$IriWrapper)) {
            self::$IriWrapper = new IriWrapper();
        }
    }

    private static function initMessenger()
    {
        if (is_null(self::$NodeMessenger)) {
            self::$NodeMessenger = new NodeMessenger();
        }
    }

    public static function processNewChunk(&$chunk)
    {
        if (self::dataNeedsAttaching($chunk)) {
            self::buildTransactionData($chunk);
            return self::sendToHookNode($chunk);
        } else {
            return 'already attached';
        }
    }

    public static function dataNeedsAttaching($request)
    {
        $command = new \stdClass();
        $command->command = "findTransactions";
        $command->addresses = array($request->address);

        BrokerNode::$iriRequestInProgress = true;

        self::initIri();
        $result = self::$IriWrapper->makeRequest($command);

        BrokerNode::$iriRequestInProgress = false;

        if (!is_null($result) && property_exists($result, 'hashes')) {
            return count($result->hashes) == 0;
        } else {
            throw new \Exception('dataNeedsAttaching failed!');
        }
    }

    public static function buildTransactionData(&$request)
    {
        $trytesToBroadcast = NULL;

        $request->value = IriData::$txValue;
        $request->tag = IriData::$oysterTag;

        $request->trytes = PrepareTransfers::buildTxTrytes($request, IriData::$oysterSeed);
        if (!is_null($request->trytes)) {
            self::getTransactionsToApprove($request);
        }

        return $request;
    }

    public static function getTransactionsToApprove(&$request)
    {
        self::initIri();

        $command = new \stdClass();
        $command->command = "getTransactionsToApprove";
        $command->depth = IriData::$depthToSearchForTxs;

        BrokerNode::$iriRequestInProgress = true;

        $result = self::$IriWrapper->makeRequest($command);

        BrokerNode::$iriRequestInProgress = false;

        if (!is_null($result)) {
            //switching trunk and branch
            //do we do this randomly or every time?
            $request->trunkTransaction = $result->branchTransaction;
            $request->branchTransaction = $result->trunkTransaction;
        } else {
            throw new \Exception('getTransactionToApprove failed!');
        }
    }

    private static function selectHookNode()
    {
        /*TODO

        remove this method and replace with Arthur's work or put Arthur's
        work in this method

        For now, we have limited nodes so we are just hard-coding nodes

        */
        return "http://165.227.79.113:250";
    }

    private static function sendToHookNode($modifiedTx)
    {
        $hookNodeUrl = self::selectHookNode();

        $tx = new \stdClass();
        $tx = $modifiedTx;
        $tx->command = 'attachToTangle';

        self::initMessenger();
        self::$NodeMessenger->sendMessageToNode($tx, $hookNodeUrl);
        self::updateHookNodeDirectory($hookNodeUrl, "request_made");

        $tx->hookNodeUrl = $hookNodeUrl;
        return $tx;
    }

    private static function updateHookNodeDirectory($currentHook, $status)
    {
        /*TODOS

        remove this method and replace with Arthur's work or put Arthur's
        work in this method
        */
        switch ($status) {
            case 'request_made':
                //we made a request
                break;
            case 'request_rejected':
                //the hooknode node declined, it doesn't know us
                //don't ask that hooknode node again
                break;
            case 'attach_completed':
                //the hooknode node says it did the POW
                break;
            case 'attach_verified':
                //we confirmed the hooknode node did the POW
                break;
            case 'attach_failed':
                //the hooknode node didn't do the POW
                //or didn't do it in time
                break;
            default:
                break;
        }
    }

    public static function verifyChunkMessageMatchesRecord($chunk)
    {
        try {
            return self::verifyChunkMatchesRecord($chunk, false);
        } catch (\Exception $e) {
            echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
        }
    }

    public static function verifyChunkMatchesRecord($chunk, $checkBranchAndTrunk = true)
    {
        $command = new \stdClass();
        $command->command = "findTransactions";
        $command->addresses = array($chunk->address);

        BrokerNode::$iriRequestInProgress = true;
        self::initIri();
        $result = self::$IriWrapper->makeRequest($command);

        BrokerNode::$iriRequestInProgress = false;

        if (!is_null($result) && property_exists($result, 'hashes') &&
            count($result->hashes) != 0) {
            $txObjects = self::getTransactionObjects($result->hashes);
            foreach ($txObjects as $key => $value) {
                if (self::chunksMatch($value, $chunk, $checkBranchAndTrunk)) {
                    return true;
                } else {
                    return false;
                }
            }
        } else {
            throw new \Exception('verifyChunkMatchesRecord failed!');
        }
    }

    public static function chunksMatch($chunkOnTangle, $chunkOnRecord, $checkBranchAndTrunk)
    {
        if ($checkBranchAndTrunk == true) {
            return self::messagesMatch($chunkOnTangle->signatureMessageFragment, $chunkOnRecord->message) &&
                $chunkOnTangle->trunkTransaction == $chunkOnRecord->trunkTransaction &&
                $chunkOnTangle->branchTransaction == $chunkOnRecord->branchTransaction;
        } else {
            return self::messagesMatch($chunkOnTangle->signatureMessageFragment, $chunkOnRecord->message);
        }
    }

    public static function messagesMatch($messageOnTangle, $messageOnRecord)
    {
        $lengthOfOriginalMessage = strlen($messageOnRecord);

        return (substr($messageOnTangle, 0, $lengthOfOriginalMessage) == $messageOnRecord) &&
            !(strlen(str_replace('9', '', substr($messageOnTangle, $lengthOfOriginalMessage))) > 0);
    }

    public static function getTransactionObjects($hashes)
    {
        $command = new \stdClass();
        $command->command = "getTrytes";
        $command->hashes = $hashes;

        BrokerNode::$iriRequestInProgress = true;
        self::initIri();
        $result = self::$IriWrapper->makeRequest($command);
        BrokerNode::$iriRequestInProgress = false;

        if (!is_null($result) && property_exists($result, 'trytes') &&
            count($result->trytes) != 0) {
            $txObjects = array();
            foreach ($result->trytes as $key => $value) {
                $txObjects[] = \Utils::transactionObject($value);
            }
            return array_reverse($txObjects);
        } else {
            throw new \Exception('getTransactionObjects failed!');
        }
    }
}
