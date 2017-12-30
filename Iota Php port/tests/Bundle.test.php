<?php
require_once('Converter.php');
require_once('Helper.php');
require_once('KerlWrapper.php');
require_once('wordsFunctionsPHP.php');
require_once('Bundle.php');

$testBundle;

$newLineChar = "<br/>";
//toggle between \n or <br/>
//depending on whether you are
//debugging in browser or terminal.

/*TODOS:
    -abstract newLineChar and newLine method to their own file so dev can change it globally in one place
    -add folder for test utilities to abstract away such files
    -create file to require_once all tests into so a dev could run all test files at once
    -create mock of Kerl
*/

//TEST VALUES
$dummyTrytes = [
    'GYOMKVTSNHVJNCNFBBAH9AAMXLPLLLROQY99QN9DLSJUHDPBLCFFAIQXZA9BKMBJCYSFHFPXAHDWZFEIZ',
    'OXJCNFHUNAHWDLKKPELTBFUCVW9KLXKOGWERKTJXQMXTKFKNWNNXYD9DMJJABSEIONOSJTTEVKVDQEWTW',
    'XQMXTKFKNWNNXYD9DMJJABSEIONOSJTTEVKVDQEWTWGYOMKVTSNHVJNCNFBBAH9AAMXLPLLLROQY99QN9',
];

$dummyValue = 'VTSNHVJNCCYSFHLKKPELTBFFPXAHDWZFENLAIQXZ9LROGYOMKVTSNHVJNQY99QNFBBAH9AAYOMKVTSNHV';


//BEGIN TESTS
newLine("Constructor");
newLine("    It should initialize an empty bundle array.");

$testBundle = new Bundle();

if ($testBundle->bundle == []) {
    newLine("        Success!");
} else {
    newLine("        FAILED!");
}

newLine("addEntry");
newLine("    It should add a certain number of transaction objects to the bundle array based on the length passed in.");

$testBundle->addEntry(
    4,
    'dummyAddress',
    $dummyValue,
    'dummyTag',
    'Four score and seven years ago',
    0 //JS functions have this index param but we don't seem to need it
);

if (count($testBundle->bundle) === 4) {
    newLine("        Success!");
} else {
    newLine("        FAILED!");
}

newLine("    It should add certain properties to each entry in the bundle.");

if (property_exists($testBundle->bundle[0], 'address') &&
    property_exists($testBundle->bundle[0], 'value') &&
    property_exists($testBundle->bundle[0], 'obsoleteTag') &&
    property_exists($testBundle->bundle[0], 'tag') &&
    property_exists($testBundle->bundle[0], 'timestamp')) {
    newLine("        Success!");
} else {
    newLine("        FAILED!");
}

newLine("    Value should be 0 for the value property of all but the first entry, and should be the value passed in for the first.");

if ($testBundle->bundle[0]->value === $dummyValue &&
    $testBundle->bundle[1]->value === 0 &&
    $testBundle->bundle[2]->value === 0 &&
    $testBundle->bundle[3]->value === 0) {
    newLine("        Success!");
} else {
    newLine("        FAILED!");
}

newLine("addTrytes");
newLine("    It should add additional properties to every bundle entry.");

$testBundle->addTrytes($dummyTrytes);

$addTrytesResult = "        Success!";

for ($i = 0; $i < count($testBundle->bundle); $i++) {
    if (!property_exists($testBundle->bundle[$i], 'signatureMessageFragment') ||
        !property_exists($testBundle->bundle[$i], 'trunkTransaction') ||
        !property_exists($testBundle->bundle[$i], 'branchTransaction') ||
        !property_exists($testBundle->bundle[$i], 'attachmentTimestamp') ||
        !property_exists($testBundle->bundle[$i], 'attachmentTimestampLowerBound') ||
        !property_exists($testBundle->bundle[$i], 'attachmentTimestampUpperBound') ||
        !property_exists($testBundle->bundle[$i], 'attachmentTimestampUpperBound') ||
        !property_exists($testBundle->bundle[$i], 'nonce')) {
        $addTrytesResult = "        FAILED!";
    }
}

newLine($addTrytesResult);

newLine("    If a signature fragment was passed in, it should use that signature fragment.");

if ($testBundle->bundle[0]->signatureMessageFragment === $dummyTrytes[0] &&
    $testBundle->bundle[1]->signatureMessageFragment === $dummyTrytes[1] &&
    $testBundle->bundle[2]->signatureMessageFragment === $dummyTrytes[2]) {
    newLine("        Success!");
} else {
    newLine("        FAILED!");
}

newLine("    If a signature fragment was NOT passed in, it should set the signature message fragment to a string of 2187 9s.");

$wholeLottaNines = '';

for ($i = 0; $i < 2187; $i++) {
    $wholeLottaNines .= '9';
}

if ($testBundle->bundle[3]->signatureMessageFragment === $wholeLottaNines) {
    newLine("        Success!");
} else {
    newLine("        FAILED!");
}

newLine("finalize");
newLine("    It create a hash and assign that hash value to the 'bundle' property of each entry.");

//FINISH THIS
////$testBundle->finalize();



//This is just so we don't have to keep adding "\n" to our print statements
function newLine($stringToPrint)
{
    global $newLineChar;

    echo $newLineChar . $stringToPrint . $newLineChar . $newLineChar;
}