#!/usr/bin/env php
<?php /* -*- coding: utf-8; indent-tabs-mode: t; tab-width: 4 -*-
vim: ts=4 noet ai */

/**
	sha3-256sum - compute SHA3-256 message digest (command-line tool)

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


use desktopd\SHA3\Sponge as SHA3;
require_once __DIR__ . '/namespaced/desktopd/SHA3/Sponge.php';

$progName = 'sha3-256sum';
$version = '2.0';
$copyrightYear = '2018';
$digestName = 'SHA3-256';
$bitNum = '256';

$args = isset ($argv) ? array_slice ($argv, 1) : array ();
$binary = false;

//require_once dirname (__FILE__) . '/SHA3.php';

function hashFile ($path = null) {
	global $progName;
	global $binary;
	
	if (!$path) {
		$path = 'php://stdin';
	} elseif (!is_file ($path)) {
		fprintf (STDERR, "%s: %s: No such file\n", $progName, $path);
		exit (1);
	} elseif (!is_readable ($path)) {
		fprintf (STDERR, "%s: %s: Permission denied\n", $progName, $path);
		exit (1);
	}
	
	$fp = fopen ($path, $binary ? 'rb' : 'r');
	$sponge = SHA3::init (SHA3::SHA3_256);
	while (!feof ($fp)) {
		$sponge->absorb (fread ($fp, 1024));
	}
	fclose ($fp);
	
	if ('php://stdin' === $path) $path = '-';
	printf ("%s %s%s\n", bin2hex ($sponge->squeeze ()), $binary ? '*' : ' '
		, $path);
}


$noArgs = true;

while (count ($args)) {
	$arg = array_shift ($args);
	switch ($arg) {
		case '--help':
			echo <<<HELP
Usage: $progName [OPTION]... [FILE]...
Print $digestName ($bitNum-bit) checksums.
With no FILE, or when FILE is -, read standard input

  -b, --binary      read in binary mode
  -t, --text        read in text mode (default)

      --help     display this help and exit.
      --version  output version information and exit.

The sums are calculated according to FIPS-202. Checking
mode is not yet implemented. See GNU Coreutils for
general ideas of sha*sum utilities.
Binary flag has an effect on DOS-based systems (i.e. Microsoft
Windows Computers (WCs))

The Desktopd project: <https://notabug.org/org/desktopd>

HELP;
			exit (0);
		
		case '--version':
			echo <<<HELP
$progName (PHP-SHA3-Streamable) $version
Copyright (C) $copyrightYear Desktopd Developers
License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Uses the PHP-SHA3-Streamable library, which is under LGPLv3+.
Modeled after GNU Coreutils' sha*sums utilities.

The Desktopd project: <https://notabug.org/org/desktopd>

HELP;
			exit (0);
		
		case '-b':
		case '--binary':
			$binary = true;
			break;
		
		case '-t':
		case '--text':
			$binary = false;
			break;
		
		default:
			hashFile ('-' === $arg ? null : $arg);
			$noArgs = false;
			break;
	}
}

if ($noArgs) {
	hashFile ();
}

