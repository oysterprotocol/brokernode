<?php

require_once("IriData.php");

class NodeMessenger
{

    private $headers = array(
        'Content-Type: application/x-www-form-urlencoded',
    );
    public $nodeUrl;
    private $userAgent = 'Codular Sample cURL Request';
    private $apiVersionHeaderString = 'X-IOTA-API-Version: ';

    public function __construct()
    {

        array_push($this->headers, $this->apiVersionHeaderString . IriData::$apiVersion);
    }

    private function validateUrl($nodeUrl)
    {

        /*
         * TODO
         * Remove this method or update it
        */

        $http = "((http)\:\/\/)"; // starts with http://
        $port = "(\:[0-9]{2,5})"; // ends with a port

        if (preg_match("/^$http/", $nodeUrl) && preg_match("/$port$/", $nodeUrl)) {
            return true;
        } else {
            throw new Exception('Invalid URL.');
        }
    }

    public function sendMessageToNode($commandObject, $nodeUrl)
    {
        $payload = http_build_query($commandObject);

        $curl = curl_init();

        curl_setopt_array($curl, array(
            CURLOPT_RETURNTRANSFER => 1,
            CURLOPT_POST => 1,
            CURLOPT_URL => $nodeUrl,
            CURLOPT_USERAGENT => $this->userAgent,
            CURLOPT_POSTFIELDS => $payload,
            CURLOPT_HTTPHEADER => $this->headers,
            CURLOPT_CONNECTTIMEOUT => 0,
            CURLOPT_TIMEOUT => 1000
        ));
        $response = json_decode(curl_exec($curl));

        if($errno = curl_errno($curl)) {
            $err_msg = curl_strerror($errno);
            curl_close($curl);
            throw new \Exception(
                "NodeMessenger Error:\n\tcURL error ({$errno}): {$err_msg}\n\tResponse: {$response}"
            );
        }

        curl_close($curl);
        return $response;
    }
}


