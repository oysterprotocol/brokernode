<?php

require_once("requests/IriWrapper.php");
require_once("requests/IriData.php");
require_once("iota-php-port/PrepareTransfers.php");
require_once("BrokerNode.php");

class ChunkProcess
{

    private $iriWrapper;
    private $currentHookNodeUrl;
    public $request;
    private $messenger;

    function __construct($request,
                         $nodeUrl = 'http://localhost:14265',
                         $apiVersion = '1.4',
                         $nodeId = 'OYSTERPEARL')
    {
        IriData::$nodeUrl = $nodeUrl;
        IriData::$apiVersion = $apiVersion;
        IriData::$oysterTag = $nodeId;
        $this->iriWrapper = new IriWrapper();
        $this->request = $request;
        $this->messenger = new NodeMessenger();
    }

    public function processNewData()
    {

        try {
            if ($this->dataNeedsAttaching($this->request)) {
                $this->getTransactionData($this->request);
                $this->sendToHookNode($this->request);
            } else {
                // move on to the next chunk
                $this->reportChunkAttached($this->request);
            }
        } catch
        (Exception $e) {
            echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
            // something went wrong during our check, do something about it
        }
    }

    private function reportChunkAttached($request)
    {
       //$messenger = new NodeMessenger();
    }

    public function setDepthToSearchForTxs($newDepth)
    {
        /* we might need to modify this while running, so exposing the means to do that
        */
        if ($newDepth >= IriData::$minDepth) {
            IriData::$depthToSearchForTxs = $newDepth;
        } else {
            IriData::$depthToSearchForTxs = IriData::$minDepth;
        }
    }

    public function setMinWeightMagnitude($newMinWeightMagnitude)
    {
        /* we might need to modify this while running, so exposing the means to do that
        */
        if ($newMinWeightMagnitude >= IriData::$minMinWeightMagnitude) {
            IriData::$minWeightMagnitude = $newMinWeightMagnitude;
        } else {
            IriData::$minWeightMagnitude = IriData::$minMinWeightMagnitude;
        }
    }

    private function selectHookNode()
    {
        /*TODO

        implement hooknode node selection

        For now, we have limited nodes so we are just hard-coding nodes

        */
        $this->currentHookNodeUrl = "http://localhost:250";
        return $this->currentHookNodeUrl;
    }

    private function sendToHookNode($modifiedTx)
    {
        $this->selectHookNode();

        $tx = new stdClass();
        $tx = $modifiedTx;
        $tx->command = 'attachToTangle';

        $this->messenger->sendMessageToNode($modifiedTx, $this->currentHookNodeUrl);
        $this->updateHookNodeDirectory($this->currentHookNodeUrl, "request_made");
        //issue request to selected hooknode node on port 250
    }

    private function updateHookNodeDirectory($currentHookNodeUrl, $status)
    {
        /*TODOS

        need to log certain info in the broker's hooknode node directory
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

    public function dataNeedsAttaching($request)
    {
        $command = new stdClass();
        $command->command = "findTransactions";
        $command->addresses = array($request->address);

        BrokerNode::$iriRequestInProgress = true;

        $result = $this->iriWrapper->makeRequest($command);

        BrokerNode::$iriRequestInProgress = false;

        if (!is_null($result) && property_exists($result, 'hashes')) {
            return count($result->hashes) == 0;
        } else {
            throw new Exception('findTransactions failed!');
        }
    }

    public function getTransactionData(&$request)
    {
        $trytesToBroadcast = NULL;

        $request->value = IriData::$txValue;
        $request->tag = IriData::$oysterTag;

        try {
            $trytes = PrepareTransfers::buildTxTrytes($request, IriData::$oysterSeed);
            $request->trytes = $trytes;
        } catch (Exception $e) {
            echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
        }

        if (!is_null($request->trytes)) {
            try {
                self::getTransactionsToApprove($request);
            } catch (Exception $e) {
                echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
            }
        }

        return $request;
    }

    private function getTransactionsToApprove(&$request)
    {
        $command = new stdClass();
        $command->command = "getTransactionsToApprove";
        $command->depth = IriData::$depthToSearchForTxs;

        BrokerNode::$iriRequestInProgress = true;

        $result = $this->iriWrapper->makeRequest($command);

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
}

