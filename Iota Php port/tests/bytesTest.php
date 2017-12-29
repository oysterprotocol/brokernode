<?php

include('Converter.php');
include('wordsFunctionsPHP.php');
require_once('Uint32Array.php');




$testTrits = [];
// echo "\n";


for ($i = 0; $i < 243; $i++) {
$testTrits[$i] = 1;
}

$wordsClass = new words();

$words = $wordsClass->trits_to_words($testTrits);

$bytes = $wordsClass->words_to_bytes($words, count($words));

//echo "\ntrits are: \n" . implode(", ", $testTrits);
echo "\nwords are: \n" . implode(", ", $words);
//echo "\nbytes are: \n" . implode(", ", $bytes);

var_dump($words->buffer);
echo "\n" . implode(", ", $words->buffer);

?>
