

<?php

require_once("../wordsFunctionsPHP.php");


$wordsToTritsTestArray = [0x00000080,0x000000020,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000];

$words = new words();

print(" ".implode($words->words_to_trits($wordsToTritsTestArray)));


$wordsArray = [0x12345678, 0xAABBCCDD];
$bytes_out;
$expected_bytes = [0xAA, 0xBB, 0xCC, 0xDD, 0x12, 0x34, 0x56, 0x78];

$ret;

$bytes_out = $words->words_to_bytes($wordsArray, count($wordsArray));

print("\nTesting words to bytes\n");
print("\n". implode(', ', $expected_bytes));
print("\n". implode(', ', $bytes_out));


$bytes_in = [0xAA, 0xBB, 0xCC, 0xDD, 0x12, 0x34, 0x56, 0x78];
$expected_words = [0x12345678, 0xAABBCCDD];
$words_out = $words->bytes_to_words($bytes_in, count($bytes_in)/4);

print("\nTesting bytes to words\n");
print("\n". implode(', ', $expected_words));
print("\n". implode(', ', $words_out));
print("\n");


?>
