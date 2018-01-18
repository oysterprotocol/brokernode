<?php

require_once("BrokerNode.php");

/*
 * TODO:
 *
 * When this file is integrated with Aaron's work, we should
 * wrap processRequest in a class and get rid of everything
 * between the start of the method and the require_once statement.
 *
 * Or we may need to delete this file altogether.
 *
 */

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

                BrokerNode::processNewChunk($request);

                break;
            case 'verifyTx':
                // client could use this to get a broker to verify
                // a tx from another broker

                try {
                    $dataIsAttached = !BrokerNode::dataNeedsAttaching($request);
                    return $dataIsAttached;
                } catch
                (Exception $e) {
                    echo "Caught exception: " . $e->getMessage();
                    // something went wrong during our check
                }

                break;
            default:
                die("UNRECOGNIZED COMMAND");
                break;
        }
    } else {
        die("NO COMMAND PROVIDED");
    }
}