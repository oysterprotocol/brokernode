<?php /* -*- coding: utf-8; indent-tabs-mode: t; tab-width: 4 -*-
vim: ts=4 noet ai */

/**
	Tests for the SHA-3 library.

	Copyright © 2016  Desktopd Developers

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
	
	@license GPL-3+
	@file
*/


// The library is under GNU LGPL 3+
require_once dirname (__FILE__) . '/SHA3.php'; // Oops, this is PHP 5.2+...

// SHA3-256("")
$sha3_256 = SHA3::init (SHA3::SHA3_256);
$sha3_256->absorb ('');
var_dump (bin2hex ($sha3_256->squeeze ())
	=== 'a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a');

// SHA3-256("")
$sha3_256 = SHA3::init (SHA3::SHA3_256);
var_dump (bin2hex ($sha3_256->squeeze ())
	=== 'a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a');

// SHA3-512("")
$sha3_512 = SHA3::init (SHA3::SHA3_512);
$sha3_512->absorb ('');
var_dump (bin2hex ($sha3_512->squeeze ())
	=== 'a69f73cca23a9ac5c8b567dc185a756e97c982164fe25859e0d1dcc1475c80a6'
		. '15b2123af1f5f94c11e3e9402c3ac558f500199d95b6d3e301758586281dcd26');

// SHA3-512("")
$sha3_512 = SHA3::init (SHA3::SHA3_512);
var_dump (bin2hex ($sha3_512->squeeze ())
	=== 'a69f73cca23a9ac5c8b567dc185a756e97c982164fe25859e0d1dcc1475c80a6'
		. '15b2123af1f5f94c11e3e9402c3ac558f500199d95b6d3e301758586281dcd26');

// SHAKE128("The quick brown fox jumps over the lazy dog", 256)
$shake128 = SHA3::init (SHA3::SHAKE128);
$shake128->absorb ('The quick brown fox jumps over the lazy dog');
var_dump (bin2hex ($shake128->squeeze (32))
	=== 'f4202e3c5852f9182a0430fd8144f0a74b95e7417ecae17db0f8cfeed0e3e66e');

// SHAKE128("The quick brown fox jumps over the lazy dog", 256)
$shake128 = SHA3::init (SHA3::SHAKE128);
$shake128->absorb ('The quick brown fox ');
$shake128->absorb ('jumps over the lazy dog');
var_dump (bin2hex ($shake128->squeeze (32))
	=== 'f4202e3c5852f9182a0430fd8144f0a74b95e7417ecae17db0f8cfeed0e3e66e');

// SHAKE256("The quick brown fox jumps over the lazy dog", 512)
$shake256 = SHA3::init (SHA3::SHAKE256);
$shake256->absorb ('The quick brown fox ');
$shake256->absorb ('jumps over the lazy dog');
var_dump (bin2hex ($shake256->squeeze (64))
	=== '2f671343d9b2e1604dc9dcf0753e5fe15c7c64a0d283cbbf722d411a0e36f6ca'
		. '1d01d1369a23539cd80f7c054b6e5daf9c962cad5b8ed5bd11998b40d5734442');

