<?php

$GLOBALS['nl'] = "\n";

class IriData
{
    public static $nodeUrl = "http://172.17.0.1:14265";
    public static $apiVersion = '1.4';

    public static $depthToSearchForTxs = 1;
    public static $minDepth = 1;
    public static $minWeightMagnitude = 14;
    public static $minMinWeightMagnitude = 14;

    public static $oysterSeed = 'OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL';
    public static $oysterTag = 'OYSTERPRL';  //will use this as the 'tag'
    public static $txValue = 0;

    public static $maxIRICallAttempts = 50;  //we need to discuss this
}

