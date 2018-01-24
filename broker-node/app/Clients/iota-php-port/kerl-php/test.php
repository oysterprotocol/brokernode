<?php
require dirname(__FILE__) . '/kerl.php';
echo "Testing 'Message' Hash To ensure Keccak is implemented correctly...\n";
$hashcheck = "0c8d6ff6e6a1cf18a0d55b20f0bca160d0d1c914a5e842f3707a25eeb20a279f6b4e83eda8e43a67697832c7f69f53ca";
$x = new Kerl();
$x->k->absorb("Message");
$hash = bin2hex($x->k->squeeze());
echo $hash;
echo "\n";
echo $hashcheck;
echo "\n";
if ($hash == $hashcheck){
    echo "SUCCESS: Keccack 384 is implemented correctly!";
}else{
    echo "\nERROR: Keccack 384 is NOT implemented correctly. Kerl will not perform properly. \nDid you edit line 51 of '/PHP-SHA3-Streamable/SHA3.php' and change '0x06' to '0x01'?";
}
echo "\n";
echo "\n";
echo "Test 1";
$trits = trytes_to_trits("EMIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH");


$kerl = new Kerl();
$kerl->absorb($trits);
$output_trits = array();
$kerl->squeeze($output_trits, 0, 243);

$x = trits_to_trytes($output_trits);
echo "\n";
echo "EMIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH";
echo "\n";
echo $x;

echo "\n";
if ($x == "EJEAOOZYSAWFPZQESYDHZCGYNSTWXUMVJOVDWUNZJXDGWCLUFGIMZRMGCAZGKNPLBRLGUNYWKLJTYEAQX"){
    echo "Test 1 Passed";
} else {
    echo "Test 1 Failed";
}



echo "\n";
echo "\n";
echo "Test 2";
$trits = trytes_to_trits("9MIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH");


$kerl = new Kerl();
$kerl->absorb($trits);
$output_trits = array();
$kerl->squeeze($output_trits, 0, 486);


$x = trits_to_trytes($output_trits);
echo "\n";
echo "9MIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH";
echo "\n";
echo $x;

echo "\n";
if ($x == "G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA"){
    echo "Test 2 Passed";
} else {
    echo "Test 2 Failed";
}

echo "\n";
echo "\n";
echo "Test 3";
$trits = trytes_to_trits("G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA");

$kerl = new Kerl();
$kerl->absorb($trits);
$output_trits = array();
$kerl->squeeze($output_trits, 0, 486);

$x = trits_to_trytes($output_trits);
echo "\n";
echo "G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA";
echo "\n";
echo $x;

echo "\n";
if ($x == "LUCKQVACOGBFYSPPVSSOXJEKNSQQRQKPZC9NXFSMQNRQCGGUL9OHVVKBDSKEQEBKXRNUJSRXYVHJTXBPDWQGNSCDCBAIRHAQCOWZEBSNHIJIGPZQITIBJQ9LNTDIBTCQ9EUWKHFLGFUVGGUWJONK9GBCDUIMAYMMQX"){
    echo "Test 3 Passed";
} else {
    echo "Test 3 Failed";
}



echo "\n";
echo "\n";
echo "Test 4";
$essence1 = [1,0,-1,1,0,-1,-1,-1,1,-1,-1,0,0,-1,-1,-1,0,0,1,0,-1,1,1,0,0,-1,0,0,-1,-1,1,1,-1,0,0,1,0,1,-1,0,0,-1,-1,0,-1,0,0,-1,-1,1,0,-1,1,-1,-1,1,0,1,1,0,0,1,1,-1,0,-1,0,-1,0,-1,-1,0,0,0,1,0,-1,0,0,-1,-1,0,1,1,-1,-1
,1,0,1,-1,0,-1,0,-1,0,1,1,-1,0,-1,1,0,1,-1,1,1,0,0,1,1,-1,1,0,0,1,0,-1,1,1,-1,1,-1,-1,-1,0,0,0,0,0,-1,0,1,-1,1,0,-1,1,-1,1,-1,0,1,0,1,1,0,1,-1,-1,1,-1,-1,0,-1,1,0,0,0,-1,0,0,0,1,-1,-1,1,1,0,-1,1,-1,1,-1,1
,1,1,1,0,-1,0,-1,0,0,1,-1,1,-1,1,1,1,-1,-1,-1,1,-1,1,-1,1,1,0,-1,0,1,-1,1,1,0,-1,-1,0,0,0,0,-1,0,-1,0,1,1,0,-1,1,1,-1,-1,1,0,1,-1,0,1,-1,1,-1,0,0,1,-1,-1,1,-1,0,-1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,1,0,1,-1,-1,0,1,0,1,1,-1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,-1,-1,1,-1,1,-1,-1,1,1,-1,1,1,-1,0,1,0,-1,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0];
$result1InTrits = [0,1,-1,1,-1,-1,-1,0,-1,-1,-1,0,-1,0,1,0,0,1,1,-1,0,1,1,-1,-1,-1,-1,-1,-1,1,0,-1,-1,1,0,-1,1,0,1,1,0,-1,1,1,0,0,1,1,1,0,0,-1,1,1,1,1,-1,1,0,1,-1,-1,0,0,1,-1,-1,-1,0,0,-1,0,-1,0,-1,0,-1,-1,
0,0,0,1,1,0,1,0,0,0,1,1,1,-1,1,1,0,1,1,-1,-1,1,0,-1,-1,1,-1,0,1,-1,1,1,0,1,0,1,0,1,0,1,0,1,-1,1,0,-1,0,-1,1,-1,1,-1,0,1,-1,-1,-1,-1,-1,0,0,-1,1,1,1,1,-1,1,1,-1,0,1,0,1,1,0,1,0,1,-1,0,0,1,-1,1,0,0,-1,-1,1,
0,0,1,0,-1,-1,0,-1,0,0,-1,1,1,-1,0,0,-1,-1,-1,-1,-1,-1,0,-1,0,0,1,0,0,-1,1,1,0,0,1,0,-1,0,0,-1,1,-1,1,1,1,1,1,0,0,1,-1,0,0,-1,0,0,0,0,1,0,0,-1,1,1,1,-1,-1,1,1,1,1,1,-1,0,0];
$result1InTrytes = 'UPQWHIYVNEOSJSDLAKVJWUWXQO9DALGJPSTUDJCJBQGHNWFMKHLCYUAEIOXFYONQIRDCZTMDUR9CFVKMZ';

$kerl = new Kerl();
$kerl->absorb($essence1);
$output_trits = array();
$kerl->squeeze($output_trits, 0, 243);

$x = trits_to_trytes($output_trits);



if ($output_trits == $result1InTrits){
    echo "\nTest 4 Trits Passed\n".trits_to_trytes($output_trits);;
} else {
    echo "\nTest 4 Trits Failed";
}

if ($x == $result1InTrytes){
    echo "\nTest 4 Trytes Passed";
} else {
    echo "\nTest 4 Trytes Failed";
}

echo "\n";
echo "\n";
echo "Test 5";
$essence2 = [1,0,-1,1,0,-1,-1,-1,1,-1,-1,0,0,-1,-1,-1,0,0,1,0,-1,1,1,0,0,-1,0,0,-1,-1,1,1,-1,0,0,1,0,1,-1,0,0,-1,-1,0,-1,0,0,-1,-1,1,0,-1,1,-1,-1,1,0,1,1,0,0,1,1,-1,0,-1,0,-1,0,-1,-1,0,0,0,1,0,-1,0,0,-1,-1,0,1,1,-1,-1
,1,0,1,-1,0,-1,0,-1,0,1,1,-1,0,-1,1,0,1,-1,1,1,0,0,1,1,-1,1,0,0,1,0,-1,1,1,-1,1,-1,-1,-1,0,0,0,0,0,-1,0,1,-1,1,0,-1,1,-1,1,-1,0,1,0,1,1,0,1,-1,-1,1,-1,-1,0,-1,1,0,0,0,-1,0,0,0,1,-1,-1,1,1,0,-1,1,-1,1,-1,1
,1,1,1,0,-1,0,-1,0,0,1,-1,1,-1,1,1,1,-1,-1,-1,1,-1,1,-1,1,1,0,-1,0,1,-1,1,1,0,-1,-1,0,0,0,0,-1,0,-1,0,1,1,0,-1,1,1,-1,-1,1,0,1,-1,0,1,-1,1,-1,0,0,1,-1,-1,1,-1,0,-1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,1,0,1,-1,-1,0,1,0,1,1,-1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,-1,-1,1,-1,1,-1,-1,1,1,-1,1,1,-1,0,1,0,-1,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0];
$result2InTrits = [0,-1,1,-1,-1,-1,1,1,0,0,0,0,1,1,1,0,1,-1,-1,-1,1,1,-1,-1,1,0,-1,0,-1,1,-1,0,1,-1,-1,0,1,0,1,0,-1,1,1,1,0,-1,0,-1,1,1,1,0,1,-1,0,0,-1,0,0,-1,0,0,-1,1,-1,0,-1,-1,0,1,-1,0,1,0,-1,-1,0,0,-1,1
,0,1,1,-1,-1,0,-1,0,0,0,-1,1,0,-1,0,-1,1,-1,1,0,1,1,-1,-1,-1,-1,-1,-1,0,0,0,0,1,-1,1,1,0,-1,0,1,1,0,1,1,0,-1,1,0,1,-1,0,-1,1,-1,0,1,-1,1,0,1,1,0,0,1,0,1,1,1,1,1,1,-1,0,1,1,1,-1,0,1,-1,1,-1,1,0,0,1,0,-1,-1
,1,-1,0,1,0,-1,-1,-1,0,0,-1,0,1,1,0,-1,0,0,0,0,1,1,0,1,0,0,-1,-1,-1,-1,1,-1,0,-1,0,-1,-1,-1,-1,0,1,0,1,0,-1,-1,1,1,-1,1,-1,-1,1,0,0,-1,0,-1,1,1,1,-1,1,0,0,-1,1,0,0,1,1,-1,-1,0];
$result2InTrytes = 'FND9MUEPSFHWJFDQMURRRYWYSZBVQ9BQGLNN9UDHJSJQYGLILMYMHTASTCNRLX9DANTXNHCEGERFVABLW';


$kerl = new Kerl();
$kerl->absorb($essence2);
$output_trits = array();
$kerl->squeeze($output_trits, 0, 243);

$x = trits_to_trytes($output_trits);


if ($output_trits == $result2InTrits){
    echo "\nTest 5 Trits Passed\n".trits_to_trytes($output_trits);
} else {
    echo "\nTest 5 Trits Failed";
}

if ($x == $result2InTrytes){
    echo "\nTest 5 Trytes Passed";
} else {
    echo "\nTest 5 Trytes Failed";
}
?>
