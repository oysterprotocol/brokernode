<?php

require_once("IriData.php");

class NodeMessenger
{

    public $headers = array(
        'Content-Type: application/json',
    );
    public $nodeUrl;
    private $userAgent = 'Codular Sample cURL Request';
    private $apiVersionHeaderString = 'X-IOTA-API-Version: ';

    public function __construct()
    {
        array_push($this->headers, $this->apiVersionHeaderString . IriData::$apiVersion);
    }

    public function sendMessageToNode($commandObject, $nodeUrl)
    {
        $payload = json_encode($commandObject);

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

        if ($errno = curl_errno($curl)) {
            $err_msg = curl_strerror($errno);
            curl_close($curl);
            throw new \Exception(
                "NodeMessenger Error:" .
                "\n\tcURL error ({$errno}): {$err_msg}" .
                "\n\tUrl: {$nodeUrl}" .
                "\n\tPayload: {$payload}" .
                "\n\tResponse: {$response}"
            );
        }

        curl_close($curl);
        return $response;
    }

    function spamHookNodes($commandObject, $nodeUrl)
    {
        $command = json_encode($commandObject);

        for ($i = 0; $i < count($nodeUrl); $i++) {

            $cmd = "curl " . $nodeUrl[$i] . " -X POST ";
            $cmd .= "-H " . "'" . $this->headers[0] . "' ";
            $cmd .= "-H " . "'" . $this->headers[1] . "' ";
            $cmd .= " -d '" . $command . "' ";
            $cmd .= " > /dev/null 2>&1 &";

            exec($cmd);
        }
    }
}


