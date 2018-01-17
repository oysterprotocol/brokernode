<?php
require_once("IriData.php");

class IriWrapper
{

    private $headers = array(
        'Content-Type: application/json',
    );
    public $nodeUrl;
    private $userAgent = 'Codular Sample cURL Request';
    private $apiVersionHeaderString = 'X-IOTA-API-Version: ';

    public function __construct()
    {
        /*$nodeUrl is expected in the following format:
            http://1.2.3.4:14265  (if using IP)
                or
            http://host:14265  (if using host)
        */
        try {
            $this->validateUrl($GLOBALS['nodeUrl']);
            array_push($this->headers, $this->apiVersionHeaderString . $GLOBALS['apiVersion']);
            $this->nodeUrl = $GLOBALS['nodeUrl'];
        } catch (Exception $e) {
            echo 'Caught exception: ' . $e->getMessage() . $GLOBALS['nl'];
        }
    }

    private function validateUrl($nodeUrl)
    {
        $http = "((http)\:\/\/)"; // starts with http://
        $port = "(\:[0-9]{2,5})"; // ends with :(port)

        /*TODOS
        add auth tokens to regex test
        */

        if (preg_match("/^$http/", $nodeUrl) && preg_match("/$port$/", $nodeUrl)) {
            return true;
        } else {
            throw new Exception('Invalid URL Format.');
        }
    }

    public function makeRequest($commandObject)
    {
        $payload = json_encode($commandObject);

        $curl = curl_init();

        curl_setopt_array($curl, array(
            CURLOPT_RETURNTRANSFER => 1,
            CURLOPT_POST => 1,
            CURLOPT_URL => $this->nodeUrl,
            CURLOPT_USERAGENT => $this->userAgent,
            CURLOPT_POSTFIELDS => $payload,
            CURLOPT_HTTPHEADER => $this->headers,
            CURLOPT_CONNECTTIMEOUT => 0,
            CURLOPT_TIMEOUT => 1000
        ));

        echo $GLOBALS['nl'];
        echo $GLOBALS['nl'] . "calling curl with command: " . $commandObject->command . $GLOBALS['nl'];
        echo $GLOBALS['nl'] . "payload: " . $payload . $GLOBALS['nl'];

        $response = json_decode(curl_exec($curl));
        curl_close($curl);

        echo $GLOBALS['nl'] . "response was: " . $GLOBALS['nl'];
        var_dump($response);

        return $response;
    }
}
