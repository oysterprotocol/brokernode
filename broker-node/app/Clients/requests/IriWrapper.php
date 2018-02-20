<?php

namespace App\Clients\requests;

class IriWrapper
{

    private static $headers = array(
        'Content-Type: application/json',
    );
    public static $nodeUrl = '';
    private static $userAgent = 'Codular Sample cURL Request';
    private static $apiVersionHeaderString = 'X-IOTA-API-Version: ';

    private static function initIri()
    {
        self::$headers[] = self::$apiVersionHeaderString . IriData::$apiVersion;
        self::$nodeUrl = IriData::$nodeUrl;
    }

    public static function makeRequest($commandObject)
    {
        if (self::$nodeUrl == '') {
            self::initIri();
        }

        $payload = json_encode($commandObject);

        $curl = curl_init();

        curl_setopt_array($curl, array(
            CURLOPT_RETURNTRANSFER => 1,
            CURLOPT_POST => 1,
            CURLOPT_URL => self::$nodeUrl,
            CURLOPT_USERAGENT => self::$userAgent,
            CURLOPT_POSTFIELDS => $payload,
            CURLOPT_HTTPHEADER => self::$headers,
            CURLOPT_CONNECTTIMEOUT => 0,
            CURLOPT_TIMEOUT => 1000
        ));

        $response = json_decode(curl_exec($curl));

        if ($errno = curl_errno($curl)) {
            $err_msg = curl_strerror($errno);
            curl_close($curl);
            throw new \Exception(
                "IriWrapper Error:" .
                "\n\tcURL error ({$errno}): {$err_msg}" .
                "\n\tUrl: " . self::$nodeUrl .
                "\n\tPayload: {$payload}" .
                "\n\tResponse: {$response}"
            );
        }

        curl_close($curl);

        return $response;
    }
}
