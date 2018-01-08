<?php

require_once("IriWrapper.php");
require_once("IriData.php");
require_once("PrepareTransfers.php");
require_once("AttachmentCheck.php");
require_once("AttachTransaction.php");
require_once("HookNode.php");

class BrokerNode
{

    private $attachmentChecker;
    private $attacher;
    private $currentHookNode;

    function __construct($nodeUrl, $apiVersion = '1.4', $nodeId = 'OYSTERPEARL')
    {
        $GLOBALS['nodeUrl'] = $nodeUrl;
        $GLOBALS['apiVersion'] = $apiVersion;
        $GLOBALS['nodeId'] = $nodeId;

        $this->attachmentChecker = new AttachmentCheck();
        $this->attacher = new AttachTransaction();
    }

    public function setDepthToSearchForTxs($newDepth)
    {
        /* depth is 18 by default, but we might
        want to change this, if so, the node
        can call this method
        */
        if ($newDepth >= $GLOBALS['minDepth']) {
            $GLOBALS['depthToSearchForTxs'] = $newDepth;
        } else {
            $GLOBALS['depthToSearchForTxs'] = $GLOBALS['minDepth'];
        }
    }

    public function setMinWeightMagnitude($newMinWeightMagnitude)
    {
        /* minWeightMagnitude is 14 by default, but we might
        want to change this, if so, the node
        can call this method
        */
        if ($newMinWeightMagnitude >= $GLOBALS['minMinWeightMagnitude']) {
            $GLOBALS['minWeightMagnitude'] = $newMinWeightMagnitude;
        } else {
            $GLOBALS['minWeightMagnitude'] = $GLOBALS['minMinWeightMagnitude'];
        }
    }

    private function selectHookNode() {
        /*TODO

        implement hook node selection

        For now, broker node will be its own hook node
        and use its own URL

        this method would need to work recursively so we don't
        keep re-selecting hook nodes that reject us

        */
        $this->currentHookNode = new HookNode($GLOBALS['nodeUrl']);
    }

    private function sendToHookNode($modifiedTx) {
        $this->selectHookNode();

        if ($this->currentHookNode->verifyRegisteredBroker($GLOBALS['nodeUrl'])) {
            $this->currentHookNode->attachTx($modifiedTx);
        } else {
            // need to pick a different hook node
            $this->sendToHookNode($modifiedTx);
        }
    }

    private function updateHookNodeDirectory($dataToLog) {
        /*TODOS

        need to log in the broker nodes' hooknode directory that
        we made this request
        */
        switch ($dataToLog) {
            case 'request_made':
                //log in directory
                break;
            case 'attach_completed':
                //log in directory
                break;
            case 'attach_verified':
                //log in directory
                break;
            case 'attach_failed':
                //log in directory
                break;
            default:
                break;
        }
    }

    public function processNewData($dataObject)
    {
        try {
            if ($this->attachmentChecker->dataNeedsAttaching($dataObject->address)) {
                $modifiedTx = $this->attacher->getTransactionData($dataObject);
                $this->sendToHookNode($modifiedTx);
                $this->updateHookNodeDirectory("request_made");
            } else {
                // move on to the next chunk
            }
        } catch (Exception $e) {
            echo "Caught exception: " . $e->getMessage() . $GLOBALS['nl'];
            // something went wrong during our check, do something about it
        }
    }
}

