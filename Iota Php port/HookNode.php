<?php

require_once("IriWrapper.php");
require_once("IriData.php");
require_once("PrepareTransfers.php");
require_once("AttachmentCheck.php");
require_once("AttachTransaction.php");
require_once("BrokerNode.php");

class HookNode
{

    private $attacher;

    public function __construct($nodeUrl, $apiVersion = '1.4', $nodeId = 'OYSTERPEARL')
    {
        $GLOBALS['nodeUrl'] = $nodeUrl;
        $GLOBALS['apiVersion'] = $apiVersion;
        $GLOBALS['nodeId'] = $nodeId;

        $this->attacher = new AttachTransaction();
    }

    public function verifyRegisteredBroker($brokerIp) {
        return true;
        /*TODO

        hook needs to check if it knows the broker
        for now, just pretend we do

        */
    }

    public function attachTx($dataObject) {
        $this->attacher->attachTx($dataObject);
    }



}