<?php /* -*- coding: utf-8; indent-tabs-mode: t; tab-width: 4 -*-
vim: ts=4 noet ai */

/**
	Example for SHA-3 usage.

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


/** Examples from https://en.wikipedia.org/wiki/SHA-3 */
var_dump (bin2hex (SHA3::init (SHA3::SHA3_256)->absorb ('')->squeeze ()));
var_dump (bin2hex (SHA3::init (SHA3::SHA3_512)->absorb ('')->squeeze ()));
var_dump (bin2hex (SHA3::init (SHA3::SHA3_224)->absorb ('')->squeeze ()));
var_dump (bin2hex (SHA3::init (SHA3::SHAKE128)
	->absorb ("The quick brown fox jumps over the lazy dog")->squeeze (32)));
var_dump (bin2hex (SHA3::init (SHA3::SHAKE128)
	->absorb ("The quick brown fox jumps over the lazy dof")->squeeze (32)));


/* Let's compare with non-PHP implementations! */

var_dump (bin2hex (SHA3::init (SHA3::SHA3_256)->absorb (pack ('H*'
	, '648a0841e8a5afb3'))->squeeze ()));
var_dump (bin2hex (SHA3::init (SHA3::SHA3_256)->absorb (pack ('H*'
	, '3e83e7c312587888af62f6cb7130ae96'))->squeeze ()));
var_dump (bin2hex (SHA3::init (SHA3::SHA3_256)->absorb (pack ('H*'
	, '24557c1ae26fa5131c2ef0aa7f265b3b795839bb6460ba4fd51bae685285cc90'))
		->squeeze ()));
var_dump (bin2hex (SHA3::init (SHA3::SHA3_256)->absorb (pack ('H*'
	, '980bdf0b69eca4943f60f3408765048c9db6543d57f507960a5543a95e5f6338'
	. '2f1cb31138dae9b7341eb67ab242e0e905bd80b0786d5f4c2049d86791c49ae5'))
		->squeeze ()));
var_dump (bin2hex (SHA3::init (SHA3::SHA3_256)->absorb (pack ('H*'
	, '33739d1f4631596954127402df40baf971a792732ad8cbdceff1909e08636e07'
	. '717da79124a71c9b351f46e5c38b8cb3ca6a2f91496ac001532eb798fe3e5505'
	. '2ff32ae1239a089b77d4db4816135e71325eb0a3d38079760bd6d3d2fbbf9a43'
	. '91836b5f48d90f4a5678437a3b960a72642a0d3c092c4f75f03d386d8fa8f1d2'))
		->squeeze ()));

var_dump (bin2hex (SHA3::init (SHA3::SHA3_512)->absorb (pack ('H*'
	, '648a0841e8a5afb3'))->squeeze ()));
var_dump (bin2hex (SHA3::init (SHA3::SHA3_512)->absorb (pack ('H*'
	, '3e83e7c312587888af62f6cb7130ae96'))->squeeze ()));
var_dump (bin2hex (SHA3::init (SHA3::SHA3_512)->absorb (pack ('H*'
	, '24557c1ae26fa5131c2ef0aa7f265b3b795839bb6460ba4fd51bae685285cc90'))
		->squeeze ()));
var_dump (bin2hex (SHA3::init (SHA3::SHA3_512)->absorb (pack ('H*'
	, '980bdf0b69eca4943f60f3408765048c9db6543d57f507960a5543a95e5f6338'
	. '2f1cb31138dae9b7341eb67ab242e0e905bd80b0786d5f4c2049d86791c49ae5'))
		->squeeze ()));
var_dump (bin2hex (SHA3::init (SHA3::SHA3_512)->absorb (pack ('H*'
	, '33739d1f4631596954127402df40baf971a792732ad8cbdceff1909e08636e07'
	. '717da79124a71c9b351f46e5c38b8cb3ca6a2f91496ac001532eb798fe3e5505'
	. '2ff32ae1239a089b77d4db4816135e71325eb0a3d38079760bd6d3d2fbbf9a43'
	. '91836b5f48d90f4a5678437a3b960a72642a0d3c092c4f75f03d386d8fa8f1d2'))
		->squeeze ()));
