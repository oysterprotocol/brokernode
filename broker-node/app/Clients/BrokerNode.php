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

    // Hack to load balance across hooknodes.
    private static $hooknode_queue = null; // errors when instantiating here?

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

    private static function getNextHooknodeIp()
    {

        $nodes = [
//          "165.227.79.113"  // test hooks
//          "104.225.221.42",
            "18.216.181.144",
            "13.59.168.153",
            "18.221.167.209",
            "18.219.8.208",
            "18.219.64.88",
            "34.215.209.25",
            "34.217.87.242",
            "52.35.133.202",
            "52.26.108.31",
            "54.200.181.188",
            "54.200.236.87",
            "54.202.4.65",
            "54.201.15.119",
            "52.41.53.151",
            "54.201.253.129",
            "54.186.146.127",
            "52.25.238.7",
            "54.202.121.124",
            "54.191.31.41",
            "54.218.125.230",
            "54.191.77.22",
            "34.213.71.222",
            "54.187.174.58",
            "34.223.224.14",
            "52.10.129.164",
            "35.183.18.203",
            "35.182.251.91",
            "35.182.249.24",
            "35.182.113.146",
            "35.182.29.53",
            "35.183.25.175",
            "35.182.238.193",
            "35.183.34.141",
            "35.183.34.61",
            "35.182.238.254",
            "35.183.23.179",
            "35.182.30.102",
            "35.182.130.104",
            "35.183.30.180",
            "35.182.255.179",
            "35.182.244.63",
            "35.182.60.28",
            "35.182.209.143",
            "35.182.208.67",
            "35.182.86.253",
            "52.69.10.141",
            "54.168.83.160",
            "13.230.200.211",
            "13.113.183.179",
            "52.198.239.38",
            "54.95.34.84",
            "52.192.179.96",
            "13.231.19.0",
            "13.114.90.229",
            "52.197.95.78",
            "13.230.165.180",
            "54.199.99.188",
            "54.95.213.25",
            "52.197.133.248",
            "54.95.4.132",
            "52.193.53.124",
            "13.231.87.152",
            "54.249.210.209",
            "54.95.23.254",
            "52.193.216.123",
            "54.174.14.237",
            "34.230.74.120",
            "34.239.104.88",
            "54.173.3.53",
            "54.89.241.181",
            "54.89.104.221",
            "54.152.16.156",
            "54.89.247.81",
            "54.211.218.168",
            "54.89.210.233",
            "54.173.11.102",
            "54.165.209.232",
            "34.238.234.90",
            "54.172.129.85",
            "54.89.219.193",
            "54.159.227.41",
            "174.129.149.121",
            "107.23.40.118",
            "54.236.165.217",
            "34.227.178.80",
            "13.127.126.78",
            "13.127.54.20",
            "13.126.92.145",
            "13.126.159.18",
            "13.127.84.34",
            "13.127.112.247",
            "13.127.107.104",
            "13.126.139.79",
            "13.127.0.27",
            "13.127.159.244",
            "35.154.5.231",
            "13.126.159.26",
            "13.127.158.170",
            "13.127.96.19",
            "13.126.44.253",
            "52.66.24.61",
            "13.126.83.90",
            "13.127.34.210",
            "13.127.125.39",
            "35.154.217.219",
            "13.124.134.6",
            "52.79.255.238",
            "13.124.182.70",
            "13.124.194.63",
            "52.78.29.167",
            "52.78.74.75",
            "13.125.140.230",
            "13.124.84.213",
            "13.125.152.222",
            "13.124.107.48",
            "13.124.101.82",
            "13.125.160.197",
            "13.125.100.211",
            "13.125.115.136",
            "13.124.254.10",
            "13.124.61.171",
            "13.125.79.128",
            "13.125.113.40",
            "13.125.70.153",
            "13.125.67.40",
            "13.56.159.224",
            "54.183.138.244",
            "13.57.9.207",
            "54.153.83.144",
            "52.53.128.165",
            "13.57.26.97",
            "54.193.108.8",
            "54.67.47.8",
            "13.57.242.118",
            "18.144.43.97",
            "18.144.64.1",
            "54.215.244.55",
            "18.144.47.61",
            "54.183.201.188",
            "13.56.156.41",
            "13.57.178.169",
            "13.56.80.147",
            "13.56.251.161",
            "18.144.60.140",
            "54.193.29.73",
            "13.229.126.222",
            "54.169.149.32",
            "13.250.99.118",
            "54.169.104.61",
            "54.255.162.242",
            "54.255.163.253",
            "13.229.198.11",
            "52.77.225.98",
            "13.250.111.229",
            "13.229.240.214",
            "13.229.129.64",
            "54.254.139.130",
            "13.250.104.13",
            "54.255.195.65",
            "54.255.172.62",
            "54.169.161.30",
            "13.250.30.90",
            "54.169.179.223",
            "13.229.103.234",
            "54.179.184.149",
            "54.153.193.193",
            "13.210.187.171",
            "52.63.211.211",
            "13.211.110.98",
            "52.65.159.1",
            "52.62.44.238",
            "52.64.199.193",
            "13.210.225.149",
            "54.153.139.176",
            "13.211.113.231",
            "54.66.232.210",
            "13.211.7.80",
            "13.54.112.238",
            "54.252.213.130",
            "13.211.89.64",
            "13.55.188.68",
            "13.210.245.189",
            "13.211.114.243",
            "54.79.97.52",
            "13.210.234.125",
            "18.196.168.191",
            "18.196.203.48",
            "18.196.187.94",
            "52.28.232.136",
            "18.196.201.98",
            "18.196.206.125",
            "35.158.44.152",
            "52.29.102.230",
            "18.195.208.251",
            "18.196.141.90",
            "52.59.108.140",
            "35.157.174.210",
            "52.59.130.226",
            "18.196.195.195",
            "35.157.56.95",
            "18.196.115.1",
            "18.196.47.126",
            "18.196.128.105",
            "18.195.244.196",
            "18.196.207.211"
        ];

        $next = array_rand($nodes);

        return $nodes[$next];
    }

    public static function processChunks(&$chunks)
    {
        if (!is_array($chunks)) {
            $chunks = array($chunks);
        }

        $addresses = array_map(function ($n) {
            return $n->address;
        }, $chunks);

        $filteredChunks = self::filterUnattachedChunks($addresses, $chunks);

        /*
         * TODO: do something with $filteredChunks->attachedChunks?
        */

        if (count($filteredChunks->unattachedChunks) != 0) {
            $chunks = $filteredChunks->unattachedChunks;
            $request = self::buildTransactionData($chunks);
            $updated_chunks = self::sendToHookNode($chunks, $request);

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
        // TODO: Use hooknodes in DB instead of this hardcode.
        // return $hooknode = HookNode::getNextReadyNode();

        return ['ip_address' => self::getNextHooknodeIp()];
    }

    private static function sendToHookNode(&$chunks, $request)
    {
        $hooknode = self::selectHookNode();
        if (empty($hooknode)) {
            return null;
        }
        $hookNodeUrl = $hooknode['ip_address'];

        $tx = new \stdClass();
        $tx = $request;
        $tx->command = 'attachToTangle';

        self::initMessenger();
        //self::$NodeMessenger->sendMessageToNode($tx, $hookNodeUrl);

        $spammedNodes = array("http://" . $hookNodeUrl . ":250/HookListener.php");   //temporary solution
        for ($i = 0; $i <= 1; $i++) {   //temporary solution
            $spammedNodes[] = "http://" . self::selectHookNode()['ip_address'] . ":250/HookListener.php";
        }

        self::$NodeMessenger->spamHookNodes($tx, $spammedNodes);  // remove this, temporary solution

        self::updateHookNodeDirectory($hookNodeUrl, "request_made");

        //record event
        self::initEventRecord();
        self::$ChunkEventsRecord->addChunkEvent("chunk_sent_to_hook", $hookNodeUrl, "todo", "todo");
        HookNode::setTimeOfLastChunk($hookNodeUrl);

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

        $addresses = array_map(function ($n) {
            return $n->address;
        }, $chunks);

        $command = new \stdClass();
        $command->command = "findTransactions";
        $command->addresses = $addresses;

        self::initIri();
        $result = self::$IriWrapper->makeRequest($command);

        $chunkResults = new \stdClass();

        $chunkResults->matchesTangle = array();
        $chunkResults->doesNotMatchTangle = array();

        if (!is_null($result) && property_exists($result, 'hashes') &&
            count($result->hashes) != 0) {

            $txObjects = self::getTransactionObjects($result->hashes);

            foreach ($chunks as $chunk) {

                $matchingTxObjects = array_filter($txObjects, function ($tx) use ($chunk, $checkBranchAndTrunk) {
                    return $tx->address == $chunk->address && self::chunksMatch($tx, $chunk, $checkBranchAndTrunk);
                });

                if (count($matchingTxObjects) != 0) {
                    $chunkResults->matchesTangle[] = $chunk;
                } else {
                    $chunkResults->doesNotMatchTangle[] = $chunk;
                }
            }

            return $chunkResults;

        } else {
            throw new \Exception('verifyChunkMatchesRecord failed!');
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
