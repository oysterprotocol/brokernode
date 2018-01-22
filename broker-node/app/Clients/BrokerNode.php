<?php

namespace App\Clients;

require_once("iota-php-port/PrepareTransfers.php");
require_once("requests/IriData.php");
require_once("requests/IriWrapper.php");
require_once("requests/NodeMessenger.php");

// This is a temporary hack to make the above required files work in this
// namespace. We can clean this up after testnet.
use \PrepareTransfers;
use \IriData;
use \IriWrapper;
use \NodeMessenger;

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

    public static function returnPrlPerGB()
    {
        /*
         * TODO:  Replace this with something
         * that takes into account the peg.
        */
        return 1;
    }

    public static function returnEthAddress()
    {
        /*
         * TODO:  This needs to call the logic
         * that Facundo is working on for issue #37
         * to return the broker node's actual eth
         * address
        */
        return "0x00F72426ccE219B5Fa1b7E4C680F4440669Ba273";
    }

    public static function processNewChunk(&$chunk)
    {
        try {
            if (self::dataNeedsAttaching($chunk)) {
                self::buildTransactionData($chunk);
                self::sendToHookNode($chunk);
            } else {
                // move on to the next chunk
                /*
                 * TODO: is there anything specific we want to do if it's
                 * already attached?
                 */
            }
        } catch
        (Exception $e) {
            echo "Caught exception: " . $e->getMessage();
            // something went wrong during our check, do something about it
        }
    }

    public static function dataNeedsAttaching($request)
    {
        $command = new stdClass();
        $command->command = "findTransactions";
        $command->addresses = array($request->address);

        BrokerNode::$iriRequestInProgress = true;

        self::initIri();
        $result = self::$IriWrapper->makeRequest($command);

        BrokerNode::$iriRequestInProgress = false;

        if (!is_null($result) && property_exists($result, 'hashes')) {
            return count($result->hashes) == 0;
        } else {
            throw new Exception('findTransactions failed!');
        }
    }

    public static function verifyChunkMatchesRecord($request)
    {
        $command = new stdClass();
        $command->command = "findTransactions";
        $command->addresses = array($request->address);

        BrokerNode::$iriRequestInProgress = true;
        self::initIri();
        $result = self::$IriWrapper->makeRequest($command);

        BrokerNode::$iriRequestInProgress = false;

        if (!is_null($result) && property_exists($result, 'hashes') &&
            count($result->hashes) != 0) {
            $txHashToCheck = end($result->hashes);
            self::getTransactionTrytes($txHashToCheck);
        } else {
            throw new Exception('findTransactions failed!');
        }
    }

    public static function getTransactionTrytes($txHash) {
        $command = new stdClass();
        $command->command = "getTrytes";
        $command->hashes = array($txHash);

        BrokerNode::$iriRequestInProgress = true;
        self::initIri();
        $result = self::$IriWrapper->makeRequest($command);

        BrokerNode::$iriRequestInProgress = false;

        if (!is_null($result) && property_exists($result, 'trytes') &&
            count($result->trytes) != 0) {
            //START HERE
        } else {
            throw new Exception('findTransactions failed!');
        }
    }

    public static function buildTransactionData(&$request)
    {
        $trytesToBroadcast = NULL;

        $request->value = IriData::$txValue;
        $request->tag = IriData::$oysterTag;

        try {
            $request->trytes = PrepareTransfers::buildTxTrytes($request, IriData::$oysterSeed);
            if (!is_null($request->trytes)) {
                self::getTransactionsToApprove($request);
            }
        } catch (Exception $e) {
            echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
        }

        return $request;
    }

    public static function getTransactionsToApprove(&$request)
    {
        self::initIri();

        $command = new stdClass();
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
            throw new Exception('getTransactionToApprove failed!');
        }
    }

    private static function selectHookNode()
    {
        /*TODO

        remove this method and replace with Arthur's work or put Arthur's
        work in this method

        For now, we have limited nodes so we are just hard-coding nodes

        */
        return "http://localhost:250";
    }

    private static function sendToHookNode($modifiedTx)
    {
        $hookNodeUrl = self::selectHookNode();

        $tx = new stdClass();
        $tx = $modifiedTx;
        $tx->command = 'attachToTangle';

        self::initMessenger();
        self::$NodeMessenger->sendMessageToNode($tx, $hookNodeUrl);
        self::updateHookNodeDirectory($hookNodeUrl, "request_made");
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


    /*
     * We don't need the methods below for anything yet.
     * Not worried about scalability now.
     */

    private static function initIfEmpty(&$objectToInit)
    {
        if (is_null($objectToInit)) {
            $objectToInit = new stdClass();
        }
    }

    private static function removeFromObject(&$objectToModify, $key)
    {
        if (isset($objectToModify->$key)) {
            unset($objectToModify->$key);
        }
    }

    public static function addChunkToAttach($chunk)
    {
        self::initIfEmpty(self::$chunksToAttach);

        $key = $chunk->chunkId;

        self::$chunksToAttach->$key = $chunk;

        self::processNewChunk($chunk);
    }

    public static function addChunkToVerify($request)
    {
        self::initIfEmpty(self::$chunksToVerify);

        $key = $request->chunkId;

        self::$chunksToVerify->$key = $request;
    }

    public static function removeFromChunksToAttach($chunk)
    {
        self::removeFromObject(self::$chunksToAttach, $chunk->chunkId);
    }

    public static function removeFromChunksToVerify($chunk)
    {
        self::removeFromObject(self::$chunksToVerify, $chunk->chunkId);
    }
}