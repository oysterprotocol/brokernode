
<?php


$trytesAlphabet = "9ABCDEFGHIJKLMNOPQRSTUVWXYZ";

$trytesTrits = [
    [ 0,  0,  0],
    [ 1,  0,  0],
    [-1,  1,  0],
    [ 0,  1,  0],
    [ 1,  1,  0],
    [-1, -1,  1],
    [ 0, -1,  1],
    [ 1, -1,  1],
    [-1,  0,  1],
    [ 0,  0,  1],
    [ 1,  0,  1],
    [-1,  1,  1],
    [ 0,  1,  1],
    [ 1,  1,  1],
    [-1, -1, -1],
    [ 0, -1, -1],
    [ 1, -1, -1],
    [-1,  0, -1],
    [ 0,  0, -1],
    [ 1,  0, -1],
    [-1,  1, -1],
    [ 0,  1, -1],
    [ 1,  1, -1],
    [-1, -1,  0],
    [ 0, -1,  0],
    [ 1, -1,  0],
    [-1,  0,  0]
];

function convtrits($trytes_param) {
	$trits = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, -1, 1, 1, -1, 1, 1, -1, 1, 0, 1, 1, 0, 1];
	
	
	global $trytesAlphabet;
	global $trytesTrits;
	print($trytes_param);
	print("in function");
	
	if(is_int( $trytes_param )){
		echo "not implemented yet";
	}
	else{
		
		for ($i = 0; $i < strlen($trytes_param); $i++) {

				$index = strpos($trytesAlphabet, substr($trytes_param, $i, 1));
				
				$trits[$i * 3] = $trytesTrits[$index][0];
				$trits[$i * 3 + 1] = $trytesTrits[$index][1];
				$trits[$i * 3 + 2] = $trytesTrits[$index][2];
				
			}
		
	}
	return $trits;
}

$testTryte = "GGGGLUVVVVJJ";
print($testTryte);

$array = convtrits($testTryte);
print(implode(" ",$array));


?>