<?php

require_once("ChunkProcess.php");
require_once("BrokerNode.php");

$req = new stdClass();

foreach ($_POST as $key => $value ) {
    $req->$key = $value;
}

$req->responseAddress = $_SERVER['REMOTE_ADDR'] . ':' . $_SERVER['REMOTE_PORT'];


processRequest($req);

function processRequest($request)
{
    if (isset($request->command)) {

        switch ($request->command) {
            case 'processNewChunk':
                // client should send this command when it has a new chunk
                // we need to process
                BrokerNode::addChunkToAttach($request);
                $process = new ChunkProcess($request);
                $process->processNewData();
                break;
            case 'verifyTx':
                // hook node should send this command when it says it is
                // done with the Tx
                BrokerNode::removeFromChunksToAttach($request);
                BrokerNode::addChunkToVerify($request);
                break;
            default:
                die("UNRECOGNIZED COMMAND");
                break;
        }
    } else {
        die("NO COMMAND PROVIDED");
    }
}