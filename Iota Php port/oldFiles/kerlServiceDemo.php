<?php

require_once('Converter.php');
require_once('wordsFunctionsPHP.php');
require_once('TypedArrayLibrary/Uint32Array.php');



//INITIALIZE KERL
$curl = curl_init();

curl_setopt_array($curl, array(
    CURLOPT_RETURNTRANSFER => 1,
    CURLOPT_URL => 'http://localhost:8081/initializeKerl',
    CURLOPT_USERAGENT => 'Codular Sample cURL Request'
));
// Send the request & save response to $resp
$key = curl_exec($curl);
print($key);
// Close request to clear up some resources
curl_close($curl);


//REFACTOR THIS SECTION (DRY)

//absorb************************
function absorb($trits, $key){
    $params = array(
        'key' => $key,
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
    // Send the request & save response to $resp
    $trits = curl_exec($curl);
    print("before trits");
    //print($trits);
    //print(json_decode($trits));
//	$trits = array_map('intval', json_decode($trits));
//	print implode(",",$trits);
    // Close request to clear up some resources
    curl_close($curl);
    return $trits;
}

//squeeze*************************
function squeeze($trits, $key, $len){
    $params = array(
        'key' => $key,
        'trits' => $trits,
        'length' => $len,
    );

    $curl = curl_init();

    curl_setopt_array($curl, array(
        CURLOPT_RETURNTRANSFER => 1,
        CURLOPT_URL => 'http://localhost:8081/squeeze',
        CURLOPT_USERAGENT => 'Codular Sample cURL Request',
        CURLOPT_POST => 1,
        CURLOPT_POSTFIELDS => http_build_query($params)
    ));

    $hash = curl_exec($curl);
    print("before trits");
    print($hash);
//	$trits = array_map('intval', json_decode($trits));
//	print("".implode(",",$trits));
    $hash = json_decode($hash);
    print($hash);
    curl_close($curl);
    return json_decode($hash);
}

//KERL TEST *******************************************************************

$chars = "GYOMKVTSNHVJNCNFBBAH9AAMXLPLLLROQY99QN9DLSJUHDPBLCFFAIQXZA9BKMBJCYSFHFPXAHDWZFEIZ";
$trytes = chars_to_trytes(str_split($chars));
$trits = Converter::trytes_to_trits($trytes);

$absorbedTrits = absorb($trits, $key);

$returnedTrits = $absorbedTrits;
//print("".implode(",",$returnedTrits));

$squeezedTrits = [];

$hash = squeeze($squeezedTrits, $key, 243);

print($hash);

// $trytes = Converter::trits_to_trytes($squeezedTrits);

// print($trytes);

// //print("".implode(",",$trytes));
// print("before chars");
// $chars2 = trytes_to_chars($trytes);
// print("\n");
// print("\n".implode("",$chars2));


//print($return);
//HELPER FUNCTIONS*************************************************************

function chars_to_trytes($charArray)
{
    print("ENTERING CHARS TO TRYTES");
    $trytesOut = array_fill(0,count($charArray),0);

    for ($i = 0; $i < count($charArray); $i++) {

        $castedChar = (((ord($charArray[$i]))+128) % 256) - 128;

        if ($castedChar == ord('9')) {
            $trytesOut[$i] = 0;
        } else if ($castedChar >= ord('N')) {
            $trytesOut[$i] = $castedChar - 64 - 27;
        } else {
            $trytesOut[$i] = $castedChar - 64;
        }

    }
    return $trytesOut;
}

function trytes_to_chars($tryteString)
{
    $tryte_to_char_mapping = "NOPQRSTUVWXYZ9ABCDEFGHIJKLM";
    $tryteCharMaps = str_split($tryte_to_char_mapping);

    $tryteArray = $tryteString;
    //print(" ".implode($tryteArray));
    $charsOut = array_fill(0,count($tryteArray),0);

    //$charsOut[27] = '\0';
    for ($i = 0; $i < count($tryteArray); $i++) {
        $charsOut[$i] = $tryteCharMaps[$tryteArray[$i] + 13];
    }


    return $charsOut;
}


?>
