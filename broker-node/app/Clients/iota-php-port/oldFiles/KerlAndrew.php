<?php

require_once('Converter.php');
require_once('wordsFunctionsPHP.php');
require_once('Uint32Array.php');
use Exception;

// define("BIT_HASH_LENGTH", 384);
// define("CURL_HASH_LENGTH", 243);
// define("BYTE_HASH_LENGTH", 384/8);
// define("RADIX", 3);
// define("MAX_TRIT_VALUE", 1);
// define("MIN_TRIT_VALUE", -1);
class Sponge {
	const SHA3_224 = 1;
	const SHA3_256 = 2;
	const SHA3_384 = 3;
	const SHA3_512 = 4;
	
	const SHAKE128 = 5;
	const SHAKE256 = 6;
	
	
	public static function init ($type = null) {
		switch ($type) {
			case self::SHA3_224: return new self (1152, 448, 0x06, 28);
			case self::SHA3_256: return new self (1088, 512, 0x06, 32);
			case self::SHA3_384: return new self (832, 768, 0x01, 48);
			case self::SHA3_512: return new self (576, 1024, 0x06, 64);
			case self::SHAKE128: return new self (1344, 256, 0x1f);
			case self::SHAKE256: return new self (1088, 512, 0x1f);
		}
		
		throw new Exception ('Invalid operation type');
	}
	
	
	/**
		Feed input to SHA-3 "sponge"
	*/
	public function absorb ($data) {
		if (self::PHASE_INPUT != $this->phase) {
			throw new Exception ('No more input accepted');
		}
		
		$rateInBytes = $this->rateInBytes;
		$this->inputBuffer .= $data;
		while (strlen ($this->inputBuffer) >= $rateInBytes) {
			list ($input, $this->inputBuffer) = array (
				substr ($this->inputBuffer, 0, $rateInBytes)
				, substr ($this->inputBuffer, $rateInBytes));
			
			$blockSize = $rateInBytes;
			for ($i = 0; $i < $blockSize; $i++) {
				$this->state[$i] = $this->state[$i] ^ $input[$i];
			}
			
			$this->state = self::keccakF1600Permute ($this->state
				, $this->native64bit);
			$this->blockSize = 0;
		}
		
		return $this;
	}
	
	/**
		Get hash output
	*/
	public function squeeze ($length = null) {
		$outputLength = $this->outputLength; // fixed length output
		if ($length && 0 < $outputLength && $outputLength != $length) {
			throw new Exception ('Invalid length');
		}
		
		if (self::PHASE_INPUT == $this->phase) {
			$this->finalizeInput ();
		}
		
		if (self::PHASE_OUTPUT != $this->phase) {
			throw new Exception ('No more output allowed');
		}
		if (0 < $outputLength) {
			$this->phase = self::PHASE_DONE;
			return $this->getOutputBytes ($outputLength);
		}
		
		$blockLength = $this->rateInBytes;
		list ($output, $this->outputBuffer) = array (
			substr ($this->outputBuffer, 0, $length)
			, substr ($this->outputBuffer, $length));
		$neededLength = $length - strlen ($output);
		$diff = $neededLength % $blockLength;
		if ($diff) {
			$readLength = (($neededLength - $diff) / $blockLength + 1)
				* $blockLength;
		} else {
			$readLength = $neededLength;
		}
		
		$read = $this->getOutputBytes ($readLength);
		$this->outputBuffer .= substr ($read, $neededLength);
		return $output . substr ($read, 0, $neededLength);
	}
	
	
	// internally used
	const PHASE_INIT = 1;
	const PHASE_INPUT = 2;
	const PHASE_OUTPUT = 3;
	const PHASE_DONE = 4;
	
	private $phase = self::PHASE_INIT;
	private $state; // byte array (string)
	private $rateInBytes; // positive integer
	private $suffix; // 8-bit unsigned integer
	private $inputBuffer = ''; // byte array (string): max length = rateInBytes
	private $outputLength = 0;
	private $outputBuffer = '';
	private $native64bit = false;
	
	
	protected function __construct ($rate, $capacity, $suffix, $length = 0) {
		if (1600 != ($rate + $capacity)) {
			throw new Error ('Invalid parameters');
		}
		if (0 != ($rate & 7)) {
			throw new Error ('Invalid rate');
		}
		
		$this->suffix = $suffix;
		$this->state = str_repeat ("\0", 200);
		$this->blockSize = 0;
		
		$this->rateInBytes = $rate / 8;
		$this->outputLength = $length;
		$this->phase = self::PHASE_INPUT;
		
		if (is_int (0x7fffffffffffffff) && is_int (-1-0x7fffffffffffffff)) {
			// 64-bit *signed* integer available, try to make use of it
			// Assuming two's complement system
			
			if (PHP_VERSION_ID >= 50603) {
				// 64-bit LE code available
				$this->native64bit = true;
			}
		}
		return;
	}
	
	protected function finalizeInput () {
		$this->phase = self::PHASE_OUTPUT;
		
		$input = $this->inputBuffer;
		$inputLength = strlen ($input);
		if (0 < $inputLength) {
			$blockSize = $inputLength;
			for ($i = 0; $i < $blockSize; $i++) {
				$this->state[$i] = $this->state[$i] ^ $input[$i];
			}
			
			$this->blockSize = $blockSize;
		}
		
		// Padding
		$rateInBytes = $this->rateInBytes;
		$this->state[$this->blockSize] = $this->state[$this->blockSize]
			^ chr ($this->suffix);
		if (($this->suffix & 0x80) != 0
			&& $this->blockSize == ($rateInBytes - 1)) {
			$this->state = self::keccakF1600Permute ($this->state
				, $this->native64bit);
		}
		$this->state[$rateInBytes - 1] = $this->state[$rateInBytes - 1] ^ "\x80";
		$this->state = self::keccakF1600Permute ($this->state
			, $this->native64bit);
	}
	
	protected function getOutputBytes ($outputLength) {
		// Squeeze
		$output = '';
		while (0 < $outputLength) {
			$blockSize = min ($outputLength, $this->rateInBytes);
			$output .= substr ($this->state, 0, $blockSize);
			$outputLength -= $blockSize;
			if (0 < $outputLength) {
				$this->state = self::keccakF1600Permute ($this->state
					, $this->native64bit);
			}
		}
		
		return $output;
	}
	
	/**
		1600-bit state version of Keccak's permutation
	*/
	protected static function keccakF1600Permute ($state, $native64bit) {
		$R = 1;
		
		if ($native64bit) {
			$pack = 'P*';
			$lanes = array_values (unpack ($pack, $state));
			
			// Notice no function calls inside the loop
			// list(), array() and unset () are not functions
			for ($round = 0; $round < 24; ++$round) {
				// θ step
				$C = array (
					$lanes[0] ^ $lanes[5] ^ $lanes[10] ^ $lanes[15] ^ $lanes[20]
					,$lanes[1] ^ $lanes[6] ^ $lanes[11] ^ $lanes[16] ^ $lanes[21]
					,$lanes[2] ^ $lanes[7] ^ $lanes[12] ^ $lanes[17] ^ $lanes[22]
					,$lanes[3] ^ $lanes[8] ^ $lanes[13] ^ $lanes[18] ^ $lanes[23]
					,$lanes[4] ^ $lanes[9] ^ $lanes[14] ^ $lanes[19] ^ $lanes[24]);
				
				$D = $C[4] ^ (($C[1] << 1) | ($C[1] >> 63) & 0x1);
				$lanes[0] ^= $D;
				$lanes[5] ^= $D;
				$lanes[10] ^= $D;
				$lanes[15] ^= $D;
				$lanes[20] ^= $D;
				
				$D = $C[0] ^ (($C[2] << 1) | ($C[2] >> 63) & 0x1);
				$lanes[1] ^= $D;
				$lanes[6] ^= $D;
				$lanes[11] ^= $D;
				$lanes[16] ^= $D;
				$lanes[21] ^= $D;
				
				$D = $C[1] ^ (($C[3] << 1) | ($C[3] >> 63) & 0x1);
				$lanes[2] ^= $D;
				$lanes[7] ^= $D;
				$lanes[12] ^= $D;
				$lanes[17] ^= $D;
				$lanes[22] ^= $D;
				
				$D = $C[2] ^ (($C[4] << 1) | ($C[4] >> 63) & 0x1);
				$lanes[3] ^= $D;
				$lanes[8] ^= $D;
				$lanes[13] ^= $D;
				$lanes[18] ^= $D;
				$lanes[23] ^= $D;
				
				$D = $C[3] ^ (($C[0] << 1) | ($C[0] >> 63) & 0x1);
				$lanes[4] ^= $D;
				$lanes[9] ^= $D;
				$lanes[14] ^= $D;
				$lanes[19] ^= $D;
				$lanes[24] ^= $D;
				
				unset ($C, $D);
				
				
				// ρ and π steps
				$current = $lanes[1];
				list ($current, $lanes[10]) = array ($lanes[10]
					, ($current << 1) | ($current >> 63) & 1);
				
				list ($current, $lanes[7]) = array ($lanes[7]
					, ($current << 3) | ($current >> 61) & 7);
				
				list ($current, $lanes[11]) = array ($lanes[11]
					, ($current << 6) | ($current >> 58) & 63);
				
				list ($current, $lanes[17]) = array ($lanes[17]
					, ($current << 10) | ($current >> 54) & 1023);
				
				list ($current, $lanes[18]) = array ($lanes[18]
					, ($current << 15) | ($current >> 49) & 32767);
				
				list ($current, $lanes[3]) = array ($lanes[3]
					, ($current << 21) | ($current >> 43) & 2097151);
				
				list ($current, $lanes[5]) = array ($lanes[5]
					, ($current << 28) | ($current >> 36) & 268435455);
				
				list ($current, $lanes[16]) = array ($lanes[16]
					, ($current << 36) | ($current >> 28) & 68719476735);
				
				list ($current, $lanes[8]) = array ($lanes[8]
					, ($current << 45) | ($current >> 19) & 35184372088831);
				
				list ($current, $lanes[21]) = array ($lanes[21]
					, ($current << 55) | ($current >> 9) & 36028797018963967);
				
				list ($current, $lanes[24]) = array ($lanes[24]
					, ($current << 2) | ($current >> 62) & 3);
				
				list ($current, $lanes[4]) = array ($lanes[4]
					, ($current << 14) | ($current >> 50) & 16383);
				
				list ($current, $lanes[15]) = array ($lanes[15]
					, ($current << 27) | ($current >> 37) & 134217727);
				
				list ($current, $lanes[23]) = array ($lanes[23]
					, ($current << 41) | ($current >> 23) & 2199023255551);
				
				list ($current, $lanes[19]) = array ($lanes[19]
					, ($current << 56) | ($current >> 8) & 72057594037927935);
				
				list ($current, $lanes[13]) = array ($lanes[13]
					, ($current << 8) | ($current >> 56) & 255);
				
				list ($current, $lanes[12]) = array ($lanes[12]
					, ($current << 25) | ($current >> 39) & 33554431);
				
				list ($current, $lanes[2]) = array ($lanes[2]
					, ($current << 43) | ($current >> 21) & 8796093022207);
				
				list ($current, $lanes[20]) = array ($lanes[20]
					, ($current << 62) | ($current >> 2) & 4611686018427387903);
				
				list ($current, $lanes[14]) = array ($lanes[14]
					, ($current << 18) | ($current >> 46) & 262143);
				
				list ($current, $lanes[22]) = array ($lanes[22]
					, ($current << 39) | ($current >> 25) & 549755813887);
				
				list ($current, $lanes[9]) = array ($lanes[9]
					, ($current << 61) | ($current >> 3) & 2305843009213693951);
				
				list ($current, $lanes[6]) = array ($lanes[6]
					, ($current << 20) | ($current >> 44) & 1048575);
				
				list ($current, $lanes[1]) = array ($lanes[1]
					, ($current << 44) | ($current >> 20) & 17592186044415);
				
				unset ($current);
				
				
				// χ step
				list (
					$lanes[0]
					, $lanes[1]
					, $lanes[2]
					, $lanes[3]
					, $lanes[4]) = array (
					$lanes[0] ^ ((~ $lanes[1]) & $lanes[2])
					, $lanes[1] ^ ((~ $lanes[2]) & $lanes[3])
					, $lanes[2] ^ ((~ $lanes[3]) & $lanes[4])
					, $lanes[3] ^ ((~ $lanes[4]) & $lanes[0])
					, $lanes[4] ^ ((~ $lanes[0]) & $lanes[1]));
				
				list (
					$lanes[5]
					, $lanes[6]
					, $lanes[7]
					, $lanes[8]
					, $lanes[9]) = array (
					$lanes[5] ^ ((~ $lanes[6]) & $lanes[7])
					, $lanes[6] ^ ((~ $lanes[7]) & $lanes[8])
					, $lanes[7] ^ ((~ $lanes[8]) & $lanes[9])
					, $lanes[8] ^ ((~ $lanes[9]) & $lanes[5])
					, $lanes[9] ^ ((~ $lanes[5]) & $lanes[6]));
				
				list (
					$lanes[10]
					, $lanes[11]
					, $lanes[12]
					, $lanes[13]
					, $lanes[14]) = array (
					$lanes[10] ^ ((~ $lanes[11]) & $lanes[12])
					, $lanes[11] ^ ((~ $lanes[12]) & $lanes[13])
					, $lanes[12] ^ ((~ $lanes[13]) & $lanes[14])
					, $lanes[13] ^ ((~ $lanes[14]) & $lanes[10])
					, $lanes[14] ^ ((~ $lanes[10]) & $lanes[11]));
				
				list (
					$lanes[15]
					, $lanes[16]
					, $lanes[17]
					, $lanes[18]
					, $lanes[19]) = array (
					$lanes[15] ^ ((~ $lanes[16]) & $lanes[17])
					, $lanes[16] ^ ((~ $lanes[17]) & $lanes[18])
					, $lanes[17] ^ ((~ $lanes[18]) & $lanes[19])
					, $lanes[18] ^ ((~ $lanes[19]) & $lanes[15])
					, $lanes[19] ^ ((~ $lanes[15]) & $lanes[16]));
				
				list (
					$lanes[20]
					, $lanes[21]
					, $lanes[22]
					, $lanes[23]
					, $lanes[24]) = array (
					$lanes[20] ^ ((~ $lanes[21]) & $lanes[22])
					, $lanes[21] ^ ((~ $lanes[22]) & $lanes[23])
					, $lanes[22] ^ ((~ $lanes[23]) & $lanes[24])
					, $lanes[23] ^ ((~ $lanes[24]) & $lanes[20])
					, $lanes[24] ^ ((~ $lanes[20]) & $lanes[21]));
				
				
				// ι step (only this step differs for each round)
				$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
				if ($R & 2) $lanes[0] ^= 1;
				$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
				if ($R & 2) $lanes[0] ^= 2;
				$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
				if ($R & 2) $lanes[0] ^= 8;
				$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
				if ($R & 2) $lanes[0] ^= 128;
				$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
				if ($R & 2) $lanes[0] ^= 32768;
				$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
				if ($R & 2) $lanes[0] ^= 2147483648;
				$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
				if ($R & 2) $lanes[0] ^= -9223372036854775807-1; // PHP_INT_MIN
			}
			
			return pack ($pack
				, $lanes[0], $lanes[1], $lanes[2], $lanes[3], $lanes[4]
				, $lanes[5], $lanes[6], $lanes[7], $lanes[8], $lanes[9]
				, $lanes[10], $lanes[11], $lanes[12], $lanes[13], $lanes[14]
				, $lanes[15], $lanes[16], $lanes[17], $lanes[18], $lanes[19]
				, $lanes[20], $lanes[21], $lanes[22], $lanes[23], $lanes[24]);
		}
		
		$lanes = str_split ($state, 8);
		
		for ($round = 0; $round < 24; ++$round) {
			// θ step
			$C = array (
				$lanes[0] ^ $lanes[5] ^ $lanes[10] ^ $lanes[15] ^ $lanes[20]
				,$lanes[1] ^ $lanes[6] ^ $lanes[11] ^ $lanes[16] ^ $lanes[21]
				,$lanes[2] ^ $lanes[7] ^ $lanes[12] ^ $lanes[17] ^ $lanes[22]
				,$lanes[3] ^ $lanes[8] ^ $lanes[13] ^ $lanes[18] ^ $lanes[23]
				,$lanes[4] ^ $lanes[9] ^ $lanes[14] ^ $lanes[19] ^ $lanes[24]);
			
			$n = $C[1];
			$D = $C[4] ^ (
				chr (((ord ($n[0]) << 1) & 0xff) ^ (ord ($n[7]) >> 7))
				.chr (((ord ($n[1]) << 1) & 0xff) ^ (ord ($n[0]) >> 7))
				.chr (((ord ($n[2]) << 1) & 0xff) ^ (ord ($n[1]) >> 7))
				.chr (((ord ($n[3]) << 1) & 0xff) ^ (ord ($n[2]) >> 7))
				.chr (((ord ($n[4]) << 1) & 0xff) ^ (ord ($n[3]) >> 7))
				.chr (((ord ($n[5]) << 1) & 0xff) ^ (ord ($n[4]) >> 7))
				.chr (((ord ($n[6]) << 1) & 0xff) ^ (ord ($n[5]) >> 7))
				.chr (((ord ($n[7]) << 1) & 0xff) ^ (ord ($n[6]) >> 7)));
			$lanes[0] ^= $D;
			$lanes[5] ^= $D;
			$lanes[10] ^= $D;
			$lanes[15] ^= $D;
			$lanes[20] ^= $D;
			
			$n = $C[2];
			$D = $C[0] ^ (
				chr (((ord ($n[0]) << 1) & 0xff) ^ (ord ($n[7]) >> 7))
				.chr (((ord ($n[1]) << 1) & 0xff) ^ (ord ($n[0]) >> 7))
				.chr (((ord ($n[2]) << 1) & 0xff) ^ (ord ($n[1]) >> 7))
				.chr (((ord ($n[3]) << 1) & 0xff) ^ (ord ($n[2]) >> 7))
				.chr (((ord ($n[4]) << 1) & 0xff) ^ (ord ($n[3]) >> 7))
				.chr (((ord ($n[5]) << 1) & 0xff) ^ (ord ($n[4]) >> 7))
				.chr (((ord ($n[6]) << 1) & 0xff) ^ (ord ($n[5]) >> 7))
				.chr (((ord ($n[7]) << 1) & 0xff) ^ (ord ($n[6]) >> 7)));
			$lanes[1] ^= $D;
			$lanes[6] ^= $D;
			$lanes[11] ^= $D;
			$lanes[16] ^= $D;
			$lanes[21] ^= $D;
			
			$n = $C[3];
			$D = $C[1] ^ (
				chr (((ord ($n[0]) << 1) & 0xff) ^ (ord ($n[7]) >> 7))
				.chr (((ord ($n[1]) << 1) & 0xff) ^ (ord ($n[0]) >> 7))
				.chr (((ord ($n[2]) << 1) & 0xff) ^ (ord ($n[1]) >> 7))
				.chr (((ord ($n[3]) << 1) & 0xff) ^ (ord ($n[2]) >> 7))
				.chr (((ord ($n[4]) << 1) & 0xff) ^ (ord ($n[3]) >> 7))
				.chr (((ord ($n[5]) << 1) & 0xff) ^ (ord ($n[4]) >> 7))
				.chr (((ord ($n[6]) << 1) & 0xff) ^ (ord ($n[5]) >> 7))
				.chr (((ord ($n[7]) << 1) & 0xff) ^ (ord ($n[6]) >> 7)));
			$lanes[2] ^= $D;
			$lanes[7] ^= $D;
			$lanes[12] ^= $D;
			$lanes[17] ^= $D;
			$lanes[22] ^= $D;
			
			$n = $C[4];
			$D = $C[2] ^ (
				chr (((ord ($n[0]) << 1) & 0xff) ^ (ord ($n[7]) >> 7))
				.chr (((ord ($n[1]) << 1) & 0xff) ^ (ord ($n[0]) >> 7))
				.chr (((ord ($n[2]) << 1) & 0xff) ^ (ord ($n[1]) >> 7))
				.chr (((ord ($n[3]) << 1) & 0xff) ^ (ord ($n[2]) >> 7))
				.chr (((ord ($n[4]) << 1) & 0xff) ^ (ord ($n[3]) >> 7))
				.chr (((ord ($n[5]) << 1) & 0xff) ^ (ord ($n[4]) >> 7))
				.chr (((ord ($n[6]) << 1) & 0xff) ^ (ord ($n[5]) >> 7))
				.chr (((ord ($n[7]) << 1) & 0xff) ^ (ord ($n[6]) >> 7)));
			$lanes[3] ^= $D;
			$lanes[8] ^= $D;
			$lanes[13] ^= $D;
			$lanes[18] ^= $D;
			$lanes[23] ^= $D;
			
			$n = $C[0];
			$D = $C[3] ^ (
				chr (((ord ($n[0]) << 1) & 0xff) ^ (ord ($n[7]) >> 7))
				.chr (((ord ($n[1]) << 1) & 0xff) ^ (ord ($n[0]) >> 7))
				.chr (((ord ($n[2]) << 1) & 0xff) ^ (ord ($n[1]) >> 7))
				.chr (((ord ($n[3]) << 1) & 0xff) ^ (ord ($n[2]) >> 7))
				.chr (((ord ($n[4]) << 1) & 0xff) ^ (ord ($n[3]) >> 7))
				.chr (((ord ($n[5]) << 1) & 0xff) ^ (ord ($n[4]) >> 7))
				.chr (((ord ($n[6]) << 1) & 0xff) ^ (ord ($n[5]) >> 7))
				.chr (((ord ($n[7]) << 1) & 0xff) ^ (ord ($n[6]) >> 7)));
			$lanes[4] ^= $D;
			$lanes[9] ^= $D;
			$lanes[14] ^= $D;
			$lanes[19] ^= $D;
			$lanes[24] ^= $D;
			
			unset ($C, $D);
			
			// ρ and π steps
			$current = $lanes[1]; // x, y
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 1;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[10]) = array ($lanes[10]
				, $current);
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 3;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[7]) = array ($lanes[7]
				, $current);
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 6;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[11]) = array ($lanes[11]
				, $current);
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 2;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[17]) = array ($lanes[17]
				, substr ($current, - 1)
					. substr ($current, 0, - 1));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 7;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[18]) = array ($lanes[18]
				, substr ($current, - 1)
					. substr ($current, 0, - 1));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 5;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[3]) = array ($lanes[3]
				, substr ($current, - 2)
					. substr ($current, 0, - 2));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 4;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[5]) = array ($lanes[5]
				, substr ($current, - 3)
					. substr ($current, 0, - 3));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 4;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[16]) = array ($lanes[16]
				, substr ($current, - 4)
					. substr ($current, 0, - 4));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 5;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[8]) = array ($lanes[8]
				, substr ($current, - 5)
					. substr ($current, 0, - 5));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 7;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[21]) = array ($lanes[21]
				, substr ($current, - 6)
					. substr ($current, 0, - 6));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 2;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[24]) = array ($lanes[24]
				, $current);
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 6;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[4]) = array ($lanes[4]
				, substr ($current, - 1)
					. substr ($current, 0, - 1));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 3;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[15]) = array ($lanes[15]
				, substr ($current, - 3)
					. substr ($current, 0, - 3));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 1;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[23]) = array ($lanes[23]
				, substr ($current, - 5)
					. substr ($current, 0, - 5));
			
			list ($current, $lanes[19]) = array ($lanes[19]
				, substr ($current, - 7)
					. substr ($current, 0, - 7));
			
			list ($current, $lanes[13]) = array ($lanes[13]
				, substr ($current, - 1)
					. substr ($current, 0, - 1));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 1;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[12]) = array ($lanes[12]
				, substr ($current, - 3)
					. substr ($current, 0, - 3));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 3;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[2]) = array ($lanes[2]
				, substr ($current, - 5)
					. substr ($current, 0, - 5));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 6;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[20]) = array ($lanes[20]
				, substr ($current, - 7)
					. substr ($current, 0, - 7));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 2;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[14]) = array ($lanes[14]
				, substr ($current, - 2)
					. substr ($current, 0, - 2));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 7;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[22]) = array ($lanes[22]
				, substr ($current, - 4)
					. substr ($current, 0, - 4));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 5;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[9]) = array ($lanes[9]
				, substr ($current, - 7)
					. substr ($current, 0, - 7));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 4;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[6]) = array ($lanes[6]
				, substr ($current, - 2)
					. substr ($current, 0, - 2));
			
			$overflow = 0x00;
			for ($i = 0; $i < 8; ++$i) {
				$a = ord ($current[$i]) << 4;
				$current[$i] = chr (0xff & $a | $overflow);
				$overflow = $a >> 8;
			}
			$current[0] = chr (ord ($current[0]) | $overflow);
			
			list ($current, $lanes[1]) = array ($lanes[1]
				, substr ($current, - 5)
					. substr ($current, 0, - 5));
			unset ($current);
			
			// χ step
			list (
				$lanes[0]
				, $lanes[1]
				, $lanes[2]
				, $lanes[3]
				, $lanes[4]) = array (
				$lanes[0] ^ ((~ $lanes[1]) & $lanes[2])
				, $lanes[1] ^ ((~ $lanes[2]) & $lanes[3])
				, $lanes[2] ^ ((~ $lanes[3]) & $lanes[4])
				, $lanes[3] ^ ((~ $lanes[4]) & $lanes[0])
				, $lanes[4] ^ ((~ $lanes[0]) & $lanes[1]));
			
			list (
				$lanes[5]
				, $lanes[6]
				, $lanes[7]
				, $lanes[8]
				, $lanes[9]) = array (
				$lanes[5] ^ ((~ $lanes[6]) & $lanes[7])
				, $lanes[6] ^ ((~ $lanes[7]) & $lanes[8])
				, $lanes[7] ^ ((~ $lanes[8]) & $lanes[9])
				, $lanes[8] ^ ((~ $lanes[9]) & $lanes[5])
				, $lanes[9] ^ ((~ $lanes[5]) & $lanes[6]));
			
			list (
				$lanes[10]
				, $lanes[11]
				, $lanes[12]
				, $lanes[13]
				, $lanes[14]) = array (
				$lanes[10] ^ ((~ $lanes[11]) & $lanes[12])
				, $lanes[11] ^ ((~ $lanes[12]) & $lanes[13])
				, $lanes[12] ^ ((~ $lanes[13]) & $lanes[14])
				, $lanes[13] ^ ((~ $lanes[14]) & $lanes[10])
				, $lanes[14] ^ ((~ $lanes[10]) & $lanes[11]));
			
			list (
				$lanes[15]
				, $lanes[16]
				, $lanes[17]
				, $lanes[18]
				, $lanes[19]) = array (
				$lanes[15] ^ ((~ $lanes[16]) & $lanes[17])
				, $lanes[16] ^ ((~ $lanes[17]) & $lanes[18])
				, $lanes[17] ^ ((~ $lanes[18]) & $lanes[19])
				, $lanes[18] ^ ((~ $lanes[19]) & $lanes[15])
				, $lanes[19] ^ ((~ $lanes[15]) & $lanes[16]));
			
			list (
				$lanes[20]
				, $lanes[21]
				, $lanes[22]
				, $lanes[23]
				, $lanes[24]) = array (
				$lanes[20] ^ ((~ $lanes[21]) & $lanes[22])
				, $lanes[21] ^ ((~ $lanes[22]) & $lanes[23])
				, $lanes[22] ^ ((~ $lanes[23]) & $lanes[24])
				, $lanes[23] ^ ((~ $lanes[24]) & $lanes[20])
				, $lanes[24] ^ ((~ $lanes[20]) & $lanes[21]));
			
			// ι step (only this step differs for each round)
			$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
			if ($R & 2) $lanes[0][0] = $lanes[0][0] ^ "\1";
			$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
			if ($R & 2) $lanes[0][0] = $lanes[0][0] ^ "\2";
			$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
			if ($R & 2) $lanes[0][0] = $lanes[0][0] ^ "\10";
			$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
			if ($R & 2) $lanes[0][0] = $lanes[0][0] ^ "\200";
			$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
			if ($R & 2) $lanes[0][1] = $lanes[0][1] ^ "\200";
			$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
			if ($R & 2) $lanes[0][3] = $lanes[0][3] ^ "\200";
			$R = (($R << 1) ^ (($R >> 7) * 0x71)) & 0xff;
			if ($R & 2) $lanes[0][7] = $lanes[0][7] ^ "\200";
		}
		
		return implode ($lanes);
	}
}


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

		print("tritslen");
		print(count($trits));
		print("endtrislen");
        for ($i = 0; $i < ($length / 242); $i++) {

            
            $words = $this->words->trits_to_words($trits);
            for ($i = 0; $i < $words->length; $i++) {
                print(".".$words[$i]);
            }
			
			
            //$bytes = $this->words->words_to_bytes($words, $words->length);

			$trytes = Converter::trits_to_trytes($trits);
			
			$chars = trytes_to_chars($trytes);
			
			
            // $num = 48;             print("ARRAY TO STRING:"); 			
			
			//print($bytes);
			//bytes to string
			print("bytes as string");
			$byteString = "".implode(",", $bytes);
			print($byteString);
			
			$string = ''.implode($chars);
			//$string = call_user_func_array("pack", array_merge(array("c*"), $bytes));
			//$string = implode(array_map("chr", $bytes));
			//$string = implode(array_map("dechex", $bytes));
			$val = sha3_update($string);
			
			print("\ntrying to get hashval\n");
			
			// //$strVal = $this->hexToStr($val);
			// //print($strVal);
			// print("\n\n");
			// print("string: ");
			
			// print("before calling chars to trytes");
			// print($val);
			// //convert from char to words
			// $n = chars_to_trytes(str_split($val));
			
			
			// $w = Converter::trytes_to_trits($n);
			
			
			// $e = $this->words->trits_to_words($w);
			
			// print($e[0]);
        }
    }

	public function hexToStr($hex){
    $string='';
    for ($i=0; $i < strlen($hex)-1; $i+=2){
        $string .= chr(hexdec($hex[$i].$hex[$i+1]));
    }
    return $string;
	}




    public function squeeze(& $trits, $offset, $length)
    {
		
		
		if ($length && (($length % self::CURL_HASH_LENGTH) !== 0)) {
            echo "Illegal length provided";
            // change this when proper error handling implemented
        }
        do {

			
			$hashCharString = sha3_finalize();
		
			$charArray = str_split($hashCharString);
			
			print("returned from finalize:");
			print($hashCharString);
			print("endhashfinalize");
			$afterSqueeze;
        
			$trytes = chars_to_trytes($charArray);
			
			
			print("\n\ntrytes to trits\n");
			print("trytes:\n");
			print("".implode(",",$trytes));
			print("\n\n");
			$trits = Converter::trytes_to_trits($trytes);
			
			print("\ntritscount:");
			print(count($trits));
			print("\n\n");
   
			$words = $this->words->trits_to_words($trits);
			
			
			print("words after squeeze");
			
			//print($words[0]);
			$bytes = $this->words->words_to_bytes($words);
			
            sha3_init();   // REPLACE WITH SHA3 EXTENSION CALLS

            for ($i = 0; $i < 48; $i++) {
                $bytes[$i] = $bytes[$i] ^ 0xFFFFFFFF;
            }

			//$string = implode(array_map("chr", $bytes));
			//$string = char
			//$string = call_user_func_array("pack", array_merge(array("C*"), $bytes));
            $afterSqueeze  = sha3_update($string);  // REPLACE WITH SHA3 EXTENSION CALLS

        } while (($length -= self::CURL_HASH_LENGTH) > 0);
		
		return $afterSqueeze;
    }
}

// C VERSION
// int chars_to_trytes(const char chars_in[], tryte_t trytes_out[], uint8_t len)
// {
    // for (uint8_t i = 0; i < len; i++) {
        // if (chars_in[i] == '9') {
            // trytes_out[i] = 0;
        // } else if ((int8_t)chars_in[i] >= 'N') {
            // trytes_out[i] = (int8_t)(chars_in[i]) - 64 - 27;
        // } else {
            // trytes_out[i] = (int8_t)(chars_in[i]) - 64;
        // }
    // }
    // return 0;
// }
		
function chars_to_trytes($charArray)
{
	print("ENTERING CHARS TO TRYTES");
	$trytesOut = array_fill(0,count($charArray),0);
	
    for ($i = 0; $i < count($charArray); $i++) {
		print("\nchar:");
		print($charArray[$i]);
		print("\n");
		$castedChar = (((ord($charArray[$i]))+128) % 256) - 128;
		print("\n".$charArray[$i]);
		print("\n".(int)$charArray[$i]);
		print("\n".$castedChar);
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


$trytes_in = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
                            -13, -12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1];

$chars = trytes_to_chars($trytes_in);
print("\nTrytes to chars test. Should be: 9ABCDEFGHIJKLMNOPQRSTUVWXYZ \n");
print(" ".implode($chars));

print("\nChars to Trytes test. Should be: {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, -13, -12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1}\n");
$trytes = chars_to_trytes($chars);

print(" ".implode($trytes));
print("\ndone\n");

$words1 = new words();

//$inputsInTrits = [1,-1,1,1,-1,0,0,-1,-1,1,1,1,-1,1,1,1,1,-1,-1,1,-1,1,0,-1,-1,-1,-1,-1,0,1,1,1,-1,1,0,1,-1,-1,-1,0,1,0,-1,-1,-1,0,-1,1,-1,1,0,-1,1,0,1,0,0,-1,0,1,0,0,0,1,0,0,1,0,0,1,1,1,0,-1,0,0,1,1,1,-1,-1,0,1,1,0,1,1,0,1,1,0,0,-1,0,-1,-1,-1,0,-1,1,-1,0,0,0,0,0,0,0,-1,0,-1,-1,-1,-1,0,0,0,1,1,0,0,1,1,1,0,-1,1,0,1,0,1,-1,-1,0,1,1,1,0,1,-1,-1,-1,1,0,0,1,1,0,1,0,0,-1,1,0,-1,1,1,0,0,0,0,1,-1,0,-1,0,-1,0,-1,0,0,1,0,0,0,0,0,-1,1,0,-1,1,1,1,1,1,-1,1,0,1,0,1,0,1,0,1,-1,0,1,0,-1,0,-1,1,-1,0,1,0,-1,1,1,-1,-1,0,-1,0,1,0,0,-1,0,1,1,1,0,-1,-1,0,-1,0,0,0,-1,1,-1,-1,1,0,0,1,-1,0,0];

$expected = 'OXJCNFHUNAHWDLKKPELTBFUCVW9KLXKOGWERKTJXQMXTKFKNWNNXYD9DMJJABSEIONOSJTTEVKVDQEWTW';

//$kerl = new Kerl();
// $kerl->initialize();


$trytes = "GYOMKVTSNHVJNCNFBBAH9AAMXLPLLLROQY99QN9DLSJUHDPBLCFFAIQXZA9BKMBJCYSFHFPXAHDWZFEIZ";

//$trytes = chars_to_trytes(str_split($trytes));

$trits = Converter::trytes_to_trits(str_split($trytes));


$b = sha3($chars);

$words = $words1->bytes_to_words(str_split($b), 12);
	

print("words:\n");

print($words[0]);

print("word count:");
print(count($words));
$trits = $words1->words_to_trits($words, 12);

$trytes = Converter::trits_to_trytes($trits);

var_dump($trytes);







print($return);



//$bytesFromHex = pack('c*',$return);
//$bytesFromHex = pack('C*',$return);
$bytesFromHex = pack('H*',$return);



$c = $words1->bytes_to_words($bytesFromHex, 12);

//print($c[0]);

$t = $words1->words_to_trits($c);

$tr = Converter::trits_to_trytes($t);

print($tr);

//$hexMsg = strToHex($chars);

//print($hexMsg);



$trytes = "EMIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH";
//$sponge = new Sponge();
$trytes = chars_to_trytes(str_split($trytes));
			
$trits = Converter::trytes_to_trits($trytes);
		
		
print("".implode(",",$trits));
   
$words = $words1->trits_to_words($trits);
			
			

$bytes = $words1->words_to_bytes($words, 12);
			
$chars = array_map("chr", $bytes);

$bin = join($chars);

$hex = bin2hex($bin);
			
    $msg = pack ('H*', $hex);
	//$result = pack ('H*', "5ee7f374973cd4bb3dc41e3081346798497ff6e36cb9352281dfe07d07fc530ca9ad8ef7aad56ef5d41be83d5e543807");
	//$length = strlen ($msg);
	
	$sponge = Sponge::init(Sponge::SHA3_384);
	$sponge->absorb ($msg);
	$b = ($sponge->squeeze());
	
	print("\n\n".$b."\n\n");

	
	print(bin2hex($b));

	$words2 = $words1->bytes_to_words(str_split($b), 12);
	

	print("words:\n");
	
	print($words2[0]);
	
	print("word count:");
	print(count($words2));
	$trits2 = $words1->words_to_trits($words2, 12);
	
	$trytes2 = Converter::trits_to_trytes($trits2);
	
	var_dump($trytes2);
	
function strToHex($string){
    $hex = '';
    for ($i=0; $i<strlen($string); $i++){
        $ord = ord($string[$i]);
        $hexCode = dechex($ord);
        $hex .= substr('0'.$hexCode, -2);
    }
    return strToUpper($hex);
}

function hexToStr($hex){
    $string='';
    for ($i=0; $i < strlen($hex)-1; $i+=2){
        $string .= chr(hexdec($hex[$i].$hex[$i+1]));
    }
    return $string;
}


// var_dump (bin2hex (Sponge::init (Sponge::SHA3_256)->absorb (pack ('H*'
	// , '980bdf0b69eca4943f60f3408765048c9db6543d57f507960a5543a95e5f6338'
	// . '2f1cb31138dae9b7341eb67ab242e0e905bd80b0786d5f4c2049d86791c49ae5'))
		// ->squeeze ()));
// //print($bytesFromHex[0]);

//$trytes = chars_to_trytes(str_split($return));

//print("".implode(",",$trytes));
// $chars = str_split($chars);

// $input = chars_to_trytes($chars);


// //
// //test this here
// //$test = new Uint32Array([1,2,500,22,11,32,43,54,1,1,2,444]);
// //print(explode(", ",$kerl->words_to_bytes($test,12)));


// print("\nCHARS TO TRYTES\n");
// print("".implode($input));

// $trits = Converter::trytes_to_trits($input);


// $newTrits = [];

// print("TRITS BEFORE ABSORB CALL ******************************************************************************");
// print(",".implode(",",$trits));


// $kerl->absorb($trits, count($trits));

// //$result1 = Converter::trits_to_trytes($trits);

// //
// //for ($i = 0; $i < 10; $i++) {
  // // print($trits[$i]);
// //}

// $kerl->squeeze($newTrits, 0, count($trits));
// print($hash);
// //
// //for ($i = 0; $i < 10; $i++) {
// //    print($newTrits[$i]);
// //}
// //
// $result = Converter::trits_to_trytes($newTrits);

// print("result");

// print($result);

// print("endresult");

// $charsAgain = trytes_to_chars(str_split($result));

// echo "\n";
// echo "\n";
// echo "\nThese two hashes should be equal:\n";
// echo " ".implode($charsAgain);
// echo "\n\n";
// echo $expected;



?>


