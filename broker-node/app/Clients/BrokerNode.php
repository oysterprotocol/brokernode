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
use App\HookNode;
use App\ChunkEvents;
use App\DataMap;

class BrokerNode
{
    public static $chunksToAttach = null;
    public static $chunksToVerify = null;

    public static $IriWrapper = null;
    public static $NodeMessenger = null;
    public static $ChunkEventsRecord = null;

    private static function initIri()
    {
        if (is_null(self::$IriWrapper)) {
            self::$IriWrapper = new IriWrapper();
        }
    }

    private static function initEventRecord()
    {
        if (is_null(self::$ChunkEventsRecord)) {
            self::$ChunkEventsRecord = new ChunkEvents();
        }
    }

    private static function initMessenger()
    {
        if (is_null(self::$NodeMessenger)) {
            self::$NodeMessenger = new NodeMessenger();
        }
    }

    public static function processChunks(&$chunks, $attachIfAlreadyAttached = false)
    {
        if (!is_array($chunks)) {
            $chunks = array($chunks);
        }

        $addresses = array();

        if (is_array($chunks[0])) {
            // in case this is coming from CheckChunkStatus which
            // will send an array of arrays
            foreach ($chunks as $key => $value) {
                $chunks[$key] = (object)$value;
                $addresses[] = $chunks[$key]->address;
            }
        } else {
            $addresses = array_map(function ($n) {
                return $n->address;
            }, $chunks);
        }

        if ($attachIfAlreadyAttached == false) {
            $filteredChunks = self::filterUnattachedChunks($addresses, $chunks);
            $chunksToSend = $filteredChunks->unattachedChunks;
            unset($filteredChunks);
        }
        else {
            $chunksToSend = $chunks;
        }

        if (count($chunksToSend) != 0) {
            $request = self::buildTransactionData($chunksToSend);
            $updated_chunks = self::sendToHookNode($chunksToSend, $request);

            DataMap::updateChunksPending($updated_chunks);

            return is_null($updated_chunks)
                ? ['hooknode_unavailable', null]
                : ['success', $updated_chunks];

        } else {
            return ['already_attached', null];
        }
    }

    public static function filterUnattachedChunks($addresses, $chunks)
    {
        $command = new \stdClass();
        $command->command = "findTransactions";
        $command->addresses = $addresses;

        self::initIri();
        $result = self::$IriWrapper->makeRequest($command);

        if (!is_null($result) && property_exists($result, 'hashes')) {

            $returnVal = new \stdClass();
            $returnVal->attachedChunks = array();
            $returnVal->unattachedChunks = array();

            if (count($result->hashes) == 0) {
                $returnVal->unattachedChunks = $chunks;
            } else {
                $txObjects = self::getTransactionObjects($result->hashes);

                $addressesOnTangle = array_map(function ($n) {
                    return $n->address;
                }, $txObjects);

                foreach ($chunks as $key => $value) {

                    in_array($value->address, $addressesOnTangle) ?
                        $returnVal->attachedChunks[] = $value :
                        $returnVal->unattachedChunks[] = $value;
                }
            }
            return $returnVal;
        } else {
            throw new \Exception(
                "BrokerNode::filterUnattachedChunks failed." .
                "\n\tIRI.findTransactions" .
                "\n\t\tcommand: {$command}" .
                "\n\t\tresult: {$result}"
            );
        }
    }

    public static function dataNeedsAttaching($request)
    {
        //this method intended to be used to check a single chunk

        $command = new \stdClass();
        $command->command = "findTransactions";
        $command->addresses = array($request->address);

        self::initIri();
        $result = self::$IriWrapper->makeRequest($command);

        if (!is_null($result) && property_exists($result, 'hashes')) {
            return count($result->hashes) == 0;
        } else {
            throw new \Exception(
                "BrokerNode::dataNeedsAttaching failed." .
                "\n\tIRI.findTransactions" .
                "\n\t\tcommand: {$command}" .
                "\n\t\tresult: {$result}"
            );
        }
    }

    private static function buildTransactionData($chunks)
    {
        if (!is_array($chunks)) {
            $chunks = array($chunks);
        }

        $trytesToBroadcast = NULL;
        $request = new \stdClass();

        foreach ($chunks as $chunk) {
            $chunk->value = IriData::$txValue;
            $chunk->tag = IriData::$oysterTag;
        }

        $request->trytes = PrepareTransfers::buildTxTrytes($chunks, IriData::$oysterSeed);
        if (!is_null($request->trytes)) {
            self::getTransactionsToApprove($request);
        }

        return $request;
    }

    private static function getTransactionsToApprove(&$request)
    {
        self::initIri();

        $command = new \stdClass();
        $command->command = "getTransactionsToApprove";
        $command->depth = IriData::$depthToSearchForTxs;

        $result = self::$IriWrapper->makeRequest($command);

        if (!is_null($result) && property_exists($result, 'branchTransaction')) {
            //switching trunk and branch
            //do we do this randomly or every time?
            $request->trunkTransaction = $result->branchTransaction;
            $request->branchTransaction = $result->trunkTransaction;
        } else {
            throw new \Exception('getTransactionToApprove failed! ' . $result->error);
        }
    }

    private static function selectHookNode()
    {
        $ready = false;

        while ($ready == false) {
            [$ready, $nextNode] = HookNode::getNextReadyNode();
        }

        return ['ip_address' => $nextNode->ip_address];
    }

    private static function sendToHookNode(&$chunks, $request)
    {
        $hooknode = self::selectHookNode();

        if (empty($hooknode)) {
            return null;
        }

        $hookNodeUrl = $hooknode['ip_address'];

        $tx = $request;
        $tx->command = 'attachToTangle';

        self::initMessenger();
        //self::$NodeMessenger->sendMessageToNode($tx, $hookNodeUrl);
      
        $spammedNodes = array("http://" . $hookNodeUrl . ":3000/");   //temporary solution

        for ($i = 0; $i <= 1; $i++) {   //temporary solution
            $spammedNodes[] = "http://" . self::selectHookNode()['ip_address'] . ":3000/";
        }

        self::$NodeMessenger->spamHookNodes($tx, $spammedNodes);  // remove this, temporary solution

        self::updateHookNodeDirectory($hookNodeUrl, "request_made");

        //record event
        self::initEventRecord();
        self::$ChunkEventsRecord->addChunkEvent("chunk_sent_to_hook", $hookNodeUrl, "todo", "todo");
        HookNode::incrementChunksProcessed($hookNodeUrl, count($chunks));

        array_walk($chunks, function ($chunk) use ($hookNodeUrl, $request) {
            $chunk->hookNodeUrl = $hookNodeUrl;
            $chunk->trunkTransaction = $request->trunkTransaction;
            $chunk->branchTransaction = $request->branchTransaction;
        });

        return $chunks;
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

    public static function verifyChunkMessagesMatchRecord($chunks)
    {
        return self::verifyChunksMatchRecord($chunks, false);
    }

    public static function verifyChunksMatchRecord($chunks, $checkBranchAndTrunk = true)
    {
        if (!is_array($chunks)) {
            $chunks = array($chunks);
        }

        $addresses = array();

        // this seems inefficient but I tried other ways and they didn't work
        // if you replace this with something better please test that the hook node
        // scores in the db still update

        if (is_array($chunks[0])) {

            // in case this is coming from CheckChunkStatus which
            // will send an array of arrays
            foreach ($chunks as $key => $value) {
                $chunks[$key] = (object)$value;
                $addresses[] = $chunks[$key]->address;
            }
        } else {
            $addresses = array_map(function ($n) {
                return $n->address;
            }, $chunks);
        }

        $command = new \stdClass();
        $command->command = "findTransactions";
        $command->addresses = $addresses;

        self::initIri();
        $result = self::$IriWrapper->makeRequest($command);

        $chunkResults = new \stdClass();

        $chunkResults->matchesTangle = array();
        $chunkResults->doesNotMatchTangle = array();
        $chunkResults->notAttached = array();

        if (!is_null($result) && property_exists($result, 'hashes') &&
            count($result->hashes) != 0) {

            $txObjects = self::getTransactionObjects($result->hashes);
            $foundAddresses = array_map(function ($n) {
                return $n->address;
            }, $txObjects);

            foreach ($chunks as $chunk) {

                $matchingTxObjects = array_filter($txObjects, function ($tx) use ($chunk, $checkBranchAndTrunk) {
                    return $tx->address == $chunk->address && self::chunksMatch($tx, $chunk, $checkBranchAndTrunk);
                });

                if (count($matchingTxObjects) != 0) {

                    $chunkResults->matchesTangle[] = $chunk;
                } else if (in_array($chunk->address, $foundAddresses)) {

                    //found address on tangle but not a match
                    $chunkResults->doesNotMatchTangle[] = $chunk;
                } else {

                    $chunkResults->notAttached[] = $chunk;
                }
            }

            return $chunkResults;

        } else if (!is_null($result) && property_exists($result, 'hashes') &&
            count($result->hashes) == 0) {

            $chunkResults->notAttached = $chunks;
            return $chunkResults;
        } else {
            $error = '';
            foreach ($result as $key => $value) {
                if (is_array($value)) {
                    $error .= $key . ": \n" . implode("\n", $value) . "\n\n";
                } else {
                    $error .= $key . ': ' . $value . "\n";
                }
                $error .= "\n";
            }

            throw new \Exception('verifyChunkMatchesRecord failed!' . $error);
        }
    }

    private static function chunksMatch($chunkOnTangle, $chunkOnRecord, $checkBranchAndTrunk)
    {
        if ($checkBranchAndTrunk == true) {
            return self::messagesMatch($chunkOnTangle->signatureMessageFragment, $chunkOnRecord->message) &&
                $chunkOnTangle->trunkTransaction == $chunkOnRecord->trunkTransaction &&
                $chunkOnTangle->branchTransaction == $chunkOnRecord->branchTransaction;
        } else {
            return self::messagesMatch($chunkOnTangle->signatureMessageFragment, $chunkOnRecord->message);
        }
    }

    private static function messagesMatch($messageOnTangle, $messageOnRecord)
    {
        $lengthOfOriginalMessage = strlen($messageOnRecord);

        return (substr($messageOnTangle, 0, $lengthOfOriginalMessage) == $messageOnRecord) &&
            !(strlen(str_replace('9', '', substr($messageOnTangle, $lengthOfOriginalMessage))) > 0);
    }

    private static function getTransactionObjects($hashes)
    {
        $command = new \stdClass();
        $command->command = "getTrytes";
        $command->hashes = $hashes;

        self::initIri();
        $result = self::$IriWrapper->makeRequest($command);

        if (!is_null($result) && property_exists($result, 'trytes') &&
            count($result->trytes) != 0) {
            $txObjects = array();
            foreach ($result->trytes as $key => $value) {
                $txObjects[] = \Utils::transactionObject($value);
            }
            return $txObjects;
        } else {
            throw new \Exception('getTransactionObjects failed!');
        }
    }
}
