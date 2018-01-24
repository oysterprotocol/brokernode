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
        (\Exception $e) {
            echo "Caught exception: " . $e->getMessage();
            // something went wrong during our check, do something about it
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

        try {
            $request->trytes = PrepareTransfers::buildTxTrytes($request, IriData::$oysterSeed);
            if (!is_null($request->trytes)) {
                self::getTransactionsToApprove($request);
            }
        } catch (\Exception $e) {
            echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
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
        return "http://localhost:250";
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

        return $hookNodeUrl;
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

    public static function verifyChunkMatchesRecord($chunk)
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
                if (self::chunksMatch($value, $chunk)) {

                    echo "CHUNK MATCHED";
                    /*TODO
                        update the status and leave the loop
                    */
                }
                else {
                    echo "CHUNK DID NOT MATCH";
                }
            }
            /*TODO
                no matches yet, respond accordingly
            */
        } else {
            throw new \Exception('verifyChunkMatchesRecord failed!');
        }
    }

    public static function chunksMatch($chunkOnTangle, $chunkOnRecord)
    {
        return self::messagesMatch($chunkOnTangle->signatureMessageFragment, $chunkOnRecord->message) &&
            $chunkOnTangle->trunkTransaction == $chunkOnRecord->trunkTransaction &&
            $chunkOnTangle->branchTransaction == $chunkOnRecord->branchTransaction;
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

    /*
     * We don't need the methods below for anything yet.
     * Not worried about scalability now.
     */

    private static function initIfEmpty(&$objectToInit)
    {
        if (is_null($objectToInit)) {
            $objectToInit = new \stdClass();
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