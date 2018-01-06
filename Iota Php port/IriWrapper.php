<?php


class IriWrapper
{

    private $headers = array(
        'Content-Type: application/json',
    );
    public $nodeUrl;
    private $userAgent = 'Codular Sample cURL Request';
    private $apiVersionHeaderString = 'X-IOTA-API-Version: ';

    public function __construct($nodeUrl, $apiVersion = '1.4')
    {
        /*$nodeUrl is expected in the following format:
            http://1.2.3.4:14265  (if using IP)
                or
            http://hostname:14265  (if using hostname)
        */
        try {
            $this->validateUrl($nodeUrl);
            array_push($this->headers, $this->apiVersionHeaderString . $apiVersion);
            $this->nodeUrl = $nodeUrl;
        } catch (Exception $e) {
            echo 'Caught exception: ',  $e->getMessage(), "\n";
        }
    }

    private function validateUrl($nodeUrl)
    {
        //need to implement this
        //if valid
        return true;

        //if not valid
        //throw new Exception('Invalid URL Format.');
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
            CURLOPT_HTTPHEADER => $this->headers
        ));

        $response = json_decode(curl_exec($curl));
        curl_close($curl);

        return $response;
    }
}
