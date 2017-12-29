<?php
include("Uint32Array.php");

//length of the uint32 array (or 'word' array) which makes up a big integer (each is 4bytes)
$INT_LENGTH = 12.0;

//equivalent length in  bytes
$BYTE_LENGTH = 48.0;

//method converts to base 3
$RADIX = 3.0;
$HALF_3 = new Uint32Array([
	0xa5ce8964,
	0x9f007669,
	0x1484504f,
	0x3ade00d9,
	0x0c24486e,
	0x50979d57,
	0x79a4c702,
	0x48bbae36,
	0xa9f6808b,
	0xaa06a805,
	0xa87fabdf,
	0x5e69ebef
]);


class words {
	/////////////////https://github.com/hikari-no-yume/TypedArrays/blob/master/src/Uint32Array.php
	
	//constructor
	public function words(){}
	

	//verbatim translation of the JS
	public function clone_uint32Array($array) {

	  $source = new Uint32Array($array);

	  return new Uint32Array($source);

	}


	//all they are doing with ta_slice is cloning.
	public function ta_slice($array){
		return new Uint32Array($array);
	}

	public function ta_reverse($array){

		$tmp = array_fill(0,$array->length,0);
		for($i = 0; $i < $array->length; $i++){
			$tmp[$i] = $array[$i];
		}
		
		$tmp = array_reverse($tmp);
		return new Uint32Array($tmp);
	}

	//negates the (unsigned) input array  
	public function bigint_not(& $array){
		for($i = 0; $i < count($array); $i++){
			$array[$i] = $this->zerofill((~$array[$i]), 0);
		}
	}

	//php implementation of >>> ref needed
	public function zerofill($a,$b) {

		if($a>=0) return $a>>$b;
		if($b==0) return (($a>>1)&0x7fffffff)*2+(($a>>$b)&1);
		else{
			return ((~$a)>>$b)^(0x7fffffff>>($b-1));
		}
	}


	public function rshift($number, $shift){
		$num = $number / pow(2,$shift);
		return $this->zerofill($num,0);
	}

	//swap endians
	public function swap32($val){
		return (($val & 0xFF) << 24) |
			(($val & 0xFF00) << 8) |
			(($val >> 8) & 0xFF00) |
			(($val >> 24) & 0xFF);
	}

	//add with carry
	public function full_add($lh, $rh, $carry){

		
		$v = $lh + $rh;
		$l = ($this->rshift($v, 32)) & 0xFFFFFFFF;
		$r = $this->zerofill(($v & 0xFFFFFFFF), 0);
		$carry1 = ($l != 0);
		
		if ($carry) {
			$v = $r + 1;
		}
		$l = ($this->rshift($v, 32)) & 0xFFFFFFFF;
		$r = $this->zerofill(($v & 0xFFFFFFFF),0);
		$carry2 = ($l != 0);

		return [$r, $carry1 || $carry2];

	}

	/// subtracts rh from base
	public function bigint_sub(& $base, $rh) {
		$noborrow = true;
		for ($i = 0; $i < $base->length; $i++) {
			$t1 = $this->zerofill(~$rh[$i], 0);
			$vc = $this->full_add($base[$i], $t1, $noborrow);

			$base[$i] = $vc[0];
			$noborrow = $vc[1];
		}

		if (!$noborrow) {
			echo "EXCEPTION - noborrow";  
		}
	}

	/// compares two (unsigned) big integers
	public function bigint_cmp($lh, $rh) {
		for ($i = count($lh); $i-- > 0;) {
			$a = $this->zerofill($lh[$i],0);
			$b = $this->zerofill($rh[$i],0);
			if ($a < $b) {
				return -1;
			} else if ($a > $b) {
				return 1;
			}
		}
		return 0;
	}

	/// adds rh to base in place
	public function bigint_add(& $base, $rh) {
		$carry = false;
		for ($i = 0; $i < count($base); $i++) {
			$vc = $this->full_add($base[$i], $rh[$i], $carry);
			$base[$i] = $vc[0];
			$carry = $vc[1];
		}
	}

	/// adds a small (i.e. <32bit) number to base
	public function bigint_add_small(& $base, $other) {
		$vc = $this->full_add($base[0], $other, false);
		$base[0] = $vc[0];
		$carry = $vc[1];

		$i = 1;
		while ($carry && ($i < $base->length)) {
		$vc = $this->full_add($base[$i], 0, $carry);
		$base[$i] = $vc[0];
		$carry = $vc[1];
		$i += 1;
		}

		return $i;
	}

	public function words_to_trits($words) {

		global $INT_LENGTH;
		global $HALF_3;
		global $RADIX;

		if (count($words) != $INT_LENGTH) {
			echo "EXCEPTION - Invalid words length";
		}

		$trits = array_fill(0, 243, 0);

		$base = new Uint32Array($words);

		$base = $this->ta_reverse($base);

		$flip_trits = false;
		
		if (($base[(int)($INT_LENGTH - 1)] >> 31) == 0) {

			// positive two's complement addition.
			$this->bigint_add($base, $HALF_3);

		} else {

			$this->bigint_not($base);
			if ($this->bigint_cmp($base, $HALF_3) > 0) {

				$this->bigint_sub($base, $HALF_3);
				$flip_trits = true;
			} else {

				$this->bigint_add_small($base, 1);
				$tmp = $this->ta_slice($HALF_3);
				$this->bigint_sub($tmp, $base);
				$base = $tmp;
			}
		}

		$rem = 0;

		for ($i = 0; $i < 242; $i++) {
			$rem = 0;
			for ($j = (int)($INT_LENGTH - 1); $j >= 0; $j--) {
				if($rem != 0){
					$firstPart = ($rem * 0xFFFFFFFF) + $rem;
				}
				else{
					$firstPart = 0;
				}

				$lhs = $firstPart + $base[$j];
				$rhs = $RADIX;

				$q = $this->zerofill(($lhs / $rhs), 0);
				$r = $this->zerofill(($lhs % $rhs), 0);

				$base[$j] = $q;
				$rem = $r;
			}
			$trits[$i] = $rem - 1;
		}

		if ($flip_trits) {
			for ($i = 0; $i < count($trits); $i++) {
				
				$trits[$i] = -$trits[$i];
			}
		}

		return $trits;
	}

	//checks if every value in an array is 0
	public function is_null_array($arr) {
		for ($i = 0; $i < count($arr); $i++) {
			if ($arr[$i] != 0) {
				return false;
				break;
			}
		}
		return true;
	}

	//in order to match the output of the iota C tests, 
	//we have to do some weird logic here  
	//In C the bytes array is an array of chars and 
	//chars are only one byte.  So when you bit shift
	//an int32 into it, you lose some of the data.  
	//In order to replicate this, we capture the results of
	//the bitshift in a string, then trim off part of it,
	//then convert the result back to decimal.
	function words_to_bytes($wordsIn, $word_len){

		$bytes_out = array_fill(0, $word_len*4, 0x0);

		//$bytes_out2 = new Uint8Array($bytes_out);

		$bits_to_shift = [24, 16, 8, 0];

		for($i = 0; $i < $word_len; $i++){

			foreach ($bits_to_shift as $index => $bits) {

				$result = dechex($wordsIn[(($word_len - 1) - $i)] >> $bits);
				$result = substr($result, $index * 2);
				$result = hexdec($result);

    				$bytes_out[$i * 4 + $index] = $result;
			}
		}
		return $bytes_out;
	}


        public function bytes_to_words($bytes_in, $word_len)
        {

            $words_out = array_fill(0, $word_len, 0);

            for ($i = 0; $i < $word_len; $i++) {
                $words_out[$i] = 0;
                $words_out[$i] |= ($bytes_in[($word_len-1-$i)*4+0] << 24) & 0xFF000000;
                $words_out[$i] |= ($bytes_in[($word_len-1-$i)*4+1] << 16) & 0x00FF0000;
                $words_out[$i] |= ($bytes_in[($word_len-1-$i)*4+2] <<  8) & 0x0000FF00;
                $words_out[$i] |= ($bytes_in[($word_len-1-$i)*4+3] <<  0) & 0x000000FF;
            }
            return $words_out;
        }

	public function trits_to_words($trits) {

		global $HALF_3;
		global $RADIX;

		print(count($trits));
		if (count($trits) != 243) {
			echo "EXCEPTION - Invalid trits length";
		}

		$base = new Uint32Array([0,0,0,0,0,0,0,0,0,0,0,0]);

		
		$tempTritsArray = array_slice($trits, 0, 242);
		
		$onlyNegativeOnes = true;
		
		for ($r = 0; $r < count($tempTritsArray); $r++) {
			if ($tempTritsArray[$r] != -1) {
				$onlyNegativeOnes = false;
			}
		}

		if ($onlyNegativeOnes) {
			$base = $this->clone_uint32Array($HALF_3);
			$this->bigint_not($base);
			$this->bigint_add_small($base, 1);
		} else {
			$size = 1;

			for ($i = count($trits)-1; $i-- > 0;) {
				$trit = $trits[$i] + 1;
				//multiply by radix
				{
					$sz = $size;
					$carry = 0;

					for ($j = 0; $j < $sz; $j++) {
						$v = ($base[$j] * $RADIX) + $carry;
						
						$carry = $this->rshift($v, 32);
						$base[$j] = $this->zerofill(($v & 0xFFFFFFFF), 0);
						
					}
					
					if ($carry > 0) {
						$base[$sz] = $carry;
						$size += 1;
					}


				}

				//addition
				{
					$sz = $this->bigint_add_small($base, $trit);
					if ($sz > $size) {
						$size = $sz;
					}
				}
			}

			if (!$this->is_null_array($base)) {
				if ($this->bigint_cmp($HALF_3, $base) <= 0) {
					$this->bigint_sub($base, $HALF_3);
				} else {
					$tmp = $this->ta_slice($HALF_3);
					$this->bigint_sub($tmp, $base);
					$this->bigint_not($tmp);
					$this->bigint_add_small($tmp, 1);
					$base = $tmp;
				}
			}
		}


		$base = $this->ta_reverse($base);
		
		for ($i = 0; $i < $base->length; $i++) {
			$base[$i] = $this->swap32($base[$i]);
		}
		return $base;
	}

}

// $wordsToTritsTestArray = [0x00000080,0x000000020,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000,0x00000000];

// $words = new words();

// print("entering words to trits test");

// $results = $words->words_to_trits($wordsToTritsTestArray);

// print(" ".implode($results));
?>
