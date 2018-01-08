<?php

require_once("IriWrapper.php");
require_once("PrepareTransfers.php");

class AttachmentCheck
{
    public function __construct()
    {
    }

    public function dataNeedsAttaching($address)
    {
        $req = new IriWrapper();

        $command = new stdClass();
        $command->command = "findTransactions";
        $command->addresses = array($address);

        $result = $req->makeRequest($command);

        if (!is_null($result) && property_exists($result, 'hashes')) {
            return count($result->hashes) == 0;
        } else {
            throw new Exception('findTransactions failed!');
        }
    }
}


