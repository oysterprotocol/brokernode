<?php /* -*- coding: utf-8; indent-tabs-mode: t; tab-width: 4 -*-
vim: ts=4 noet ai */

/**
	Streamable SHA-3 for PHP 5.2+, with no lib/ext dependencies!

	Copyright © 2018  Desktopd Developers

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Lesser General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Lesser General Public License for more details.

	You should have received a copy of the GNU Lesser General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.

	@license LGPL-3+
	@file
*/


namespace desktopd\SHA3;
use Exception;

/**
	SHA-3 (FIPS-202) for PHP strings (byte arrays) (PHP 5.3+)
	PHP 7.0 computes SHA-3 about 4 times faster than PHP 5.2 - 5.6 (on x86_64)

	Based on the reference implementations, which are under CC-0
	Reference: http://keccak.noekeon.org/

	This uses PHP's native byte strings. Supports 32-bit as well as 64-bit
	systems. Also for LE vs. BE systems.
*/
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
			case self::SHA3_384: return new self (832, 768, 0x06, 48);
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
	
	
	public function __construct ($rate, $capacity, $suffix, $length = 0) {
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

