<?php

namespace App\Clients;

use GuzzleHttp\Client;
use GuzzleHttp\Psr7\Request;

// TODO: Use real URL from env variable.
const ROOT_URL = 'https://jsonplaceholder.typicode.com/';
const HEADERS = ['Content-Type' => 'application/json'];

class HookNode
{

    public static function processNewChunk($address, $message, $chunk_idx, $cb) {
        $client = new Client(['base_uri' => ROOT_URL]);
        return $client->requestAsync('POST', 'posts/1', [
            'command' => 'processNewChunk',
            'address' => $address,
            'message' => $message,
            'chunkId' => $chunk_idx
        ]);
    }

}