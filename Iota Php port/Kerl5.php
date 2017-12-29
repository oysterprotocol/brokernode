<?php

require_once('Converter.php');
require_once('wordsFunctionsPHP.php');
require_once('Uint32Array.php');


// define("BIT_HASH_LENGTH", 384);
// define("CURL_HASH_LENGTH", 243);
// define("BYTE_HASH_LENGTH", 384/8);
// define("RADIX", 3);
// define("MAX_TRIT_VALUE", 1);
// define("MIN_TRIT_VALUE", -1);

class Kerl
{
    const BIT_HASH_LENGTH = 384;
    const CURL_HASH_LENGTH = 243;
    const BYTE_HASH_LENGTH = 384 / 4;
    const RADIX = 3;
    const MAX_TRIT_VALUE = 1;
    const MIN_TRIT_VALUE = -1;

    public $words;
    public $k;
    private $byte_state = [];
    private $trit_state = [];

    public function __construct()
    {
        //$this->k = sha3("Message", 384);
        $this->words = new words();
	echo 'made it to construct in kerl';
        //$this->k = new stdClass();
        $retVal = sha3_init();
        print($retVal);
        echo "\nreturn val of init:" . $retVal . "\n";
	echo "made it past sha call in kerl constructor\n";
        //print("TYPE::\n");
        // print(gettype($this->k));
    }

    public function initialize($state = null)
    {
echo "made it to initalize in kerl";

        $this->byte_state = [];
        $this->trit_state = [];

        if (!$state) {
            $this->trit_state = array_fill(0, self::CURL_HASH_LENGTH * self::RADIX, 0);
            return;
        }

        // initialize empty trit state
        for ($i = 0; $i < self::CURL_HASH_LENGTH * self::RADIX; $i++) {
            $this->trit_state[$i] = $state[$i];
        }
    }

    public function reset()
    {
        //$this->k = sha3("Message", 384);
        //$this->k = new stdClass();
        sha3_init();        // REPLACE WITH SHA3 EXTENSION CALLS
    }

    public function absorb($trits, $length)
    {
		print("Absorbing");
		
		for ($i = 0; $i < 10; $i++) {
                print($trits[$i]);
            }

        for ($i = 0; $i < ($length / 243); $i++) {
            // Last trit to zero
            //$trits = array_fill(0,243,0);

            //$bytes = array_fill(0, 48, 0);
            //$packed = pack("c*", ...$bytes);
            //$bytes = new Uint8Array($bytes);

            //$words = array_fill(0,12,0);

            //memcpy(trits, &trits_in[i*243], 242);
 
            // First, convert to bytes
            //int32_t words[12];
            //unsigned char bytes[48];
            
            //$words = $this->words->trits_to_words($trits);
            
            //$bytes = $this->words->words_to_bytes($words, count($words));

            $trytes = Converter::trits_to_trytes($trits);
			
            sha3_update($trytes, 48);
       //     echo "\nType of bytes is\n";
	     //     echo gettype($bytes);
	     //     echo "\nvalue of bytes\n";
       //     echo implode(', ', $bytes);
			
			//$packed = pack("C*",$bytes);
			//print($packed);
			
       //     $num = 48;
       //     print("ARRAY TO STRING:");
			
			
		//	print($bytes);
			//bytes to string
		//	$byteString = implode("", $bytes);
		//	print($byteString);
		//	$str2 = "sdfdsfsfsfsds";
		//	print($str2);
			
			//var_dump($bytes);
		//	print("csdscdsc");
		//	$strlen = strlen($byteString);
			//$file = fopen('message.txt', 'w');
			//fwrite($file, $string);
 
   //   $val = sha3_update($byteString, 48);
		//	print("trying to get hashval");
		//	print($val);
			
      //$val = explode("_", $val);
      
      
			//might need that weird char to bytes method in the c implementation
			//$out = $this->words->bytes_to_words($val, count($val)/4);
      
      
      
      
			//echo "bytes array after coming back from keccak_update and being made into an array again:\n";
      //echo implode(", ", 
			
			//does not match JS version
			//for($i = 0; $i < 10; $i++)
			//{
			   //print($out[$i]."\n");
			//}
			// save the col	umn headers
			// fputcsv($file, array('string'));
			 
			// // Sample data. This can be fetched from mysql too
			// $data = array(
				// array($str2)
			// );
			 
			// // save each row of the data
			// foreach ($data as $row)
			// {
				// fputcsv($file, $row);
			// }
			 
			// Close the file
			//fclose($file);
						
			
        }
    }

    public function js_absorb($trits, $offset, $length) {

        if ($length && (($length % self::CURL_HASH_LENGTH) !== 0)) {
            echo "Illegal length provided";
            // change this when proper error handling implemented
        }

        do {
            $limit = ($length < self::CURL_HASH_LENGTH ? $length : self::CURL_HASH_LENGTH);

            $trit_state = array_slice($trits, $offset, $offset + $limit);
            $offset += $limit;

            // convert trit state to words
            $wordsToAbsorb = $this->words->trits_to_words($trit_state);
            $bytesToAbsorb = $this->words->words_to_bytes($wordsToAbsorb);

            // absorb the trit state as bytes
            sha3_update($bytesToAbsorb, 48);

//            this.k.update( // REPLACE WITH SHA3 EXTENSION CALLS
//                CryptoJS.lib.WordArray.create(wordsToAbsorb));  // REPLACE WITH SHA3 EXTENSION CALLS

        } while (($length -= self::CURL_HASH_LENGTH) > 0);
    }

    //think word len is 12
    // public function words_to_bytes($wordsIn, $word_len){

    // print($wordsIn->length);

    // $bytes_out = array_fill(0,48,0);

    // for($i = 0; $i < $word_len; $i++){
    // $bytes_out[$i*4+0] = ($wordsIn[($word_len-1)-$i] >> 24);
    // $bytes_out[$i*4+1] = ($wordsIn[($word_len-1)-$i] >> 16);
    // $bytes_out[$i*4+2] = ($wordsIn[($word_len-1)-$i] >> 8);
    // $bytes_out[$i*4+3] = ($wordsIn[($word_len-1)-$i] >> 0);

    // }

    // return $bytes_out;

    // }


    public function squeeze(& $trits, $length)
    {
        if ($length && (($length % self::CURL_HASH_LENGTH) !== 0)) {
            echo "Illegal length provided";
            // change this when proper error handling implemented
        }
        do {

            $bytes = array_fill(0, 48, 0);
            $packed = pack("c*", ...$bytes);

//            print("bytes");
//            print($bytes[0]);
//            print("lenbytes");
//            print(count($bytes));
//            print("packed");
//            print($packed[0]);
            // get the hash digest
            // get the hash digest
            //$kCopy = $this->k;  // REPLACE WITH SHA3 EXTENSION CALLS

            //print($kCopy);

            $result = sha3_finalize();  // REPLACE WITH SHA3 EXTENSION CALLS
            return $result;

            //print($bytes[0]);

            //Convert bytes to words

            //$convertedWords = $this->words->bytes_to_words($bytes, count($bytes) / 4);
            //words to trits
            //$trit_state = $this->words->words_to_trits($convertedWords);

            //$i = 0;
            //$limit = ($length < self::CURL_HASH_LENGTH ? $length : self::CURL_HASH_LENGTH);

            //while ($i < $limit) {

           //     $trits[$offset++] = $trit_state[$i++];
           // }

           // $this->reset();   // REPLACE WITH SHA3 EXTENSION CALLS

          //  for ($i = 0; $i < 48; $i++) {
          //      $bytes[$i] = $bytes[$i] ^ 0xFFFFFFFF;
          //  }

            //sha3_update($bytes, 48);  // REPLACE WITH SHA3 EXTENSION CALLS

        } while (($length -= self::CURL_HASH_LENGTH) > 0);
    }
}


$input = 'GYOMKVTSNHVJNCNFBBAH9AAMXLPLLLROQY99QN9DLSJUHDPBLCFFAIQXZA9BKMBJCYSFHFPXAHDWZFEIZ';

$inputsInTrits = [1,-1,1,1,-1,0,0,-1,-1,1,1,1,-1,1,1,1,1,-1,-1,1,-1,1,0,-1,-1,-1,-1,-1,0,1,1,1,-1,1,0,1,-1,-1,-1,0,1,0,-1,-1,-1,0,-1,1,-1,1,0,-1,1,0,1,0,0,-1,0,1,0,0,0,1,0,0,1,0,0,1,1,1,0,-1,0,0,1,1,1,-1,-1,0,1,1,0,1,1,0,1,1,0,0,-1,0,-1,-1,-1,0,-1,1,-1,0,0,0,0,0,0,0,-1,0,-1,-1,-1,-1,0,0,0,1,1,0,0,1,1,1,0,-1,1,0,1,0,1,-1,-1,0,1,1,1,0,1,-1,-1,-1,1,0,0,1,1,0,1,0,0,-1,1,0,-1,1,1,0,0,0,0,1,-1,0,-1,0,-1,0,-1,0,0,1,0,0,0,0,0,-1,1,0,-1,1,1,1,1,1,-1,1,0,1,0,1,0,1,0,1,-1,0,1,0,-1,0,-1,1,-1,0,1,0,-1,1,1,-1,-1,0,-1,0,1,0,0,-1,0,1,1,1,0,-1,-1,0,-1,0,0,0,-1,1,-1,-1,1,0,0,1,-1,0,0];

$expected = 'OXJCNFHUNAHWDLKKPELTBFUCVW9KLXKOGWERKTJXQMXTKFKNWNNXYD9DMJJABSEIONOSJTTEVKVDQEWTW';

$kerl = new Kerl();
$kerl->initialize();

//test this here
//$test = new Uint32Array([1,2,500,22,11,32,43,54,1,1,2,444]);
//print(explode(", ",$kerl->words_to_bytes($test,12)));




$trits = Converter::trytes_to_trits($input);
$newTrits = [];

//print(count($trits));
$kerl->absorb($trits, count($trits));
$kerl->squeeze($newTrits, count($trits));

$result = Converter::trits_to_trytes($newTrits);

//
for ($i = 0; $i < 10; $i++) {
   print($trits[$i]);
}

//$kerl->squeeze($newTrits, 0, count($trits));
//
//for ($i = 0; $i < 10; $i++) {
//    print($newTrits[$i]);
//}
//
//$result = Converter::trits_to_trytes($newTrits);


echo "\n";
echo "\n";
echo "\nThese two hashes should be equal:\n";
echo $result;
echo "\n\n";
echo $expected;



?>


