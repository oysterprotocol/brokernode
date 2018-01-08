<?php

require_once("IriWrapper.php");
require_once("IriData.php");
require_once("PrepareTransfers.php");
require_once("AttachmentCheck.php");
require_once("AttachTransaction.php");

class Node
{

    private $attachmentChecker;
    private $attacher;

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
        /* minWeightMagnitude is 18 by default, but we might
        want to change this, if so, the node
        can call this method
        */
        if ($newMinWeightMagnitude >= $GLOBALS['minMinWeightMagnitude']) {
            $GLOBALS['minWeightMagnitude'] = $newMinWeightMagnitude;
        } else {
            $GLOBALS['minWeightMagnitude'] = $GLOBALS['minMinWeightMagnitude'];
        }
    }

    public function dataIsUnattached($address)
    {
        $result = null;
        $i = 0;

        while ($result == NULL && $i <= $GLOBALS['maxIRICallAttempts']) { //how many times should we attempt the check?
            $result = $this->attachmentChecker->attachmentCheck($address);
            $i++;
        }

        if (is_null($result) || property_exists($result, 'error') || is_null($result->hashes)) {
            return 'error;';
            // something went wrong during our check
        } else if (count($result->hashes) != 0) {
            return 'attached';
        } else {
            return 'unattached';
        }
    }

    public function processNewData($dataObject)
    {

        $this->attacher->attachTx($dataObject);
        //remove this, do proper checks


//        if ($this->dataIsUnattached($dataObject->address) == 'unattached') {
//            $this->attacher->attachTx($dataObject);
//            // attach the data or hand off to hook node to attach
//        } else if ($this->dataIsUnattached($dataObject->address) == 'attached') {
//            // get next chunk of data
//        } else {
//            // something went wrong during our check, do something about it
//        }
    }
}

$node = new Node('http://104.5.45.19:14265');

$transactionObject = new stdClass();

$transactionObject->address = '9YSTERPRLOYSTERPRLOYSTMMMRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL';
$transactionObject->message = 'SOMEFAKEDATATHATWEWANTTOPASSINFOROURTESTSBOYISUREHOPETHISWORKS';

$node->processNewData($transactionObject);
