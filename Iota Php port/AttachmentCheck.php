<?php

require_once("IriWrapper.php");
require_once("PrepareTransfers.php");

class AttachmentCheck
{
    public function __construct()
    {
    }

    public function attachmentCheck($address)
    {
        $req = new IriWrapper();

        $command = new stdClass();
        $command->command = "findTransactions";
        $command->addresses = array($address);

        $result = $req->makeRequest($command);

        return $result;
    }
}


