#!/usr/bin/env php
<?php /* -*- coding: utf-8; indent-tabs-mode: t; tab-width: 4 -*-
vim: ts=4 noet ai */

use desktopd\SHA3\Sponge as SHA3;

require __DIR__ . '/namespaced/desktopd/SHA3/Sponge.php';

$length = 1024 * 1024; // 1MiB
$data = str_repeat ("\0", $length);

$start = microtime ();
$sponge = SHA3::init (SHA3::SHA3_256);
$sponge->absorb ($data);
$hash = $sponge->squeeze ();
$end = microtime ();

$start = explode (' ', $start);
$end = explode (' ', $end);
printf ("%d Bytes in %.6f seconds\n%s\n"
	, $length
	, ($end[1] - $start[1]) + ($end[0] - $start[0])
	, bin2hex ($hash));

