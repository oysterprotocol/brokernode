<?php

//remove this once the tests aren't in this file anymore
require_once('Converter.php');


/*This file is to serve as a "wrapper" for the node service until we can get a pure php
keccak384 algorithm working for our needs. The methods in this KerlWrapper file have
all the same params as the methods in the JS version, and it even has the class name
"Kerl".  The goal is to use this file for now, and when we have a pure PHP solution,
we should be able to remove this file and put in the new one without having to modify
how any of our other files call the Kerl.
*/


class Kerl
{

    public $kerlKey;


    public function __construct()
    {
        //This method doesn't do anything yet,
        //but when we get a pure php Kerl working
        //we will need to construct it, so this
        //method is here so that our files which
        //use Kerl have a "constructor" to call,
        //so that when we implement a pure php kerl
        //we will not have to change any other files
        //that use it.
    }


    //INITIALIZE KERL
    //initialize************************
    public function initialize()
    {
        $curl = curl_init();

        curl_setopt_array($curl, array(
            CURLOPT_RETURNTRANSFER => 1,
            CURLOPT_URL => 'http://localhost:8081/initializeKerl',
            CURLOPT_USERAGENT => 'Codular Sample cURL Request'
        ));

        // Send the request
        $this->kerlKey = curl_exec($curl);
        echo $this->kerlKey;

        // Close request to clear up some resources
        curl_close($curl);
    }

    //reset************************
    public function reset()
    {
        $params = array(
            'key' => $this->kerlKey,
        );

        $curl = curl_init();

        curl_setopt_array($curl, array(
            CURLOPT_RETURNTRANSFER => 1,
            CURLOPT_URL => 'http://localhost:8081/reset',
            CURLOPT_USERAGENT => 'Codular Sample cURL Request',
            CURLOPT_POST => 1,
            CURLOPT_POSTFIELDS => http_build_query($params)
        ));

        // Send the request
        curl_exec($curl);

        curl_close($curl);
    }

    //absorb************************
    public function absorb($trits, $offset, $length)
        //we don't currently use these last two params
        //but we will when we get a pure PHP kerl ready,
        //so adding the params to the method signature
        //so we won't have to change other files in the future.
    {
        $params = array(
            'key' => $this->kerlKey,
            'trits' => $trits,
        );

        $curl = curl_init();

        curl_setopt_array($curl, array(
            CURLOPT_RETURNTRANSFER => 1,
            CURLOPT_URL => 'http://localhost:8081/absorb',
            CURLOPT_USERAGENT => 'Codular Sample cURL Request',
            CURLOPT_POST => 1,
            CURLOPT_POSTFIELDS => http_build_query($params)
        ));

        // Send the request
        curl_exec($curl);

        curl_close($curl);
    }

    //squeeze*************************
    public function squeeze(&$trits, $offset, $length)
        //we don't currently use the offset param but
        //presumably we will in the future, so adding it
        //in now so we won't have to change other files
        //in the future.
    {
        $params = array(
            'key' => $this->kerlKey,
            'trits' => $trits,
            'length' => $length,
        );

        $curl = curl_init();

        curl_setopt_array($curl, array(
            CURLOPT_RETURNTRANSFER => 1,
            CURLOPT_URL => 'http://localhost:8081/squeeze',
            CURLOPT_USERAGENT => 'Codular Sample cURL Request',
            CURLOPT_POST => 1,
            CURLOPT_POSTFIELDS => http_build_query($params)
        ));

        $returnedTrits = curl_exec($curl);
        $returnedTrits = json_decode($returnedTrits);

        $trits = $returnedTrits;

        curl_close($curl);
    }
}

//KERL TEST *******************************************************************
/*
$trytes = "GYOMKVTSNHVJNCNFBBAH9AAMXLPLLLROQY99QN9DLSJUHDPBLCFFAIQXZA9BKMBJCYSFHFPXAHDWZFEIZ";

$trits = Converter::trytes_to_trits($trytes);

$kerl = new Kerl();

$kerl->initialize();
$kerl->absorb($trits, 0, count($trits));

$squeezedTrits = [];

$kerl->squeeze($squeezedTrits, 0, count($trits));

$trytes = Converter::trits_to_trytes($squeezedTrits);

echo "\n Returned trits, converted to trytes: \n" . $trytes . "\n";
echo "\n Expected result:  \n";
echo "OXJCNFHUNAHWDLKKPELTBFUCVW9KLXKOGWERKTJXQMXTKFKNWNNXYD9DMJJABSEIONOSJTTEVKVDQEWTW";
echo "\n";
*/
?>