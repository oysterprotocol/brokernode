<?php

require_once("requests/IriWrapper.php");
require_once("requests/IriData.php");
require_once("iota-php-port/PrepareTransfers.php");
require_once("ChunkProcess.php");

class BrokerNode
{
    public static $nodePort = 350;
    public static $hookNodePort = 250;
    public static $chunksToAttach = null;
    public static $chunksToVerify = null;

    public static $iriRequestInProgress = false;


    public static function addChunkToAttach($chunk)
    {

        if (is_null(self::$chunksToAttach)) {
            self::$chunksToAttach = new stdClass();
        }

        $key = $chunk->chunkId;

        self::$chunksToAttach->$key = $chunk;

        $chunkProcess = new ChunkProcess($chunk);

        $chunkProcess->processNewData();
    }

    public static function addChunkToVerify($chunk)
    {

        if (is_null(self::$chunksToVerify)) {
            self::$chunksToVerify = new stdClass();
        }

        $key = $chunk->chunkId;

        self::$chunksToVerify->$key = $chunk;
    }

    public static function removeFromChunksToAttach($chunk)
    {
        $key = $chunk->chunkId;

        if (isset(self::$chunksToAttach->$key)) {
            unset(self::$chunksToAttach->$key);
        }
    }

    public static function removeFromChunksToVerify($chunk)
    {
        $key = $chunk->chunkId;

        if (isset(self::$chunksToVerify->$key)) {
            unset(self::$chunksToVerify->$key);
        }
    }
}