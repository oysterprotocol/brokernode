#!/usr/bin/env php
<?php /* -*- coding: utf-8; indent-tabs-mode: t; tab-width: 4 -*-
vim: ts=4 noet ai */

use desktopd\SHA3\Sponge as SHA3;

require __DIR__ . '/namespaced/desktopd/SHA3/Sponge.php';


// 1024 bytes
const DATA = '
4v9WYVupNPLoX6SLlFaBbmKrcxnf5tHcln6UT1V5O/wqdjLsue1w80yuCsqs/vmtWoHRVFLE7pzu
iyeWOhXJ5EFLq2AazMwENB2KsliR94XZhX15tGIVduC4M2Dp+grh76hIHYqxCC0KK52YzF2JKVI9
EP1G5wTRCDSPRttk5nadineOevtXgh6hZMionuqkCaAasRpNCdOU6kxBLXeA4wKBAbUpR9HEC7YU
xo0REqD018D+CZlBJVZq2Gb241T84gnxbXwfrOzgjc9+J+AZzj5W1PJct+AulNdA+wDkG0PnQg0R
2DYRqk46S6l3vJ5Y4e/ind8f8rXKpXtSqllkxgxigrqvTvTBxRBTAjoJIPjc0GekNJc1++Zsvdp0
llYK0O8l5YrhHZWJfRWImR/J2GCfifxFZ79RR9BBt9zdy12ni5uk4xLOE/tSNlqK1LRu65UUL7xa
lDLWD82SSxTeSN5INEXoMJ4S8TwMpNNSaUy39A7qj0jNcKoum4+9snEP7YttIdM7u54gVsW1qoIj
eej07gVATWFj/amJdLCodv7aMvxG/MTTc2DSfRJTCa8Z4w4cW1yT9rXPT0jCO2nitq0eIVUd4xg0
toZnY5tonoGlVVG38pXSuGV21ocr0BjTiQweZJ5x2OdiABN5sdvqrbwBV0VkK/gMLxPnFIMMM94G
oRLvcEXlvNLdKlxmgWMmDgi7O34yUpns3fGI9kUMhfo9GoSlqpguzHhPQh+hrrnVLQEjJCkNTbm9
R6DUX/JUW/pyQ9th3sHNiQUWJ+PWTby2Fj8f9FFQvlWY9FTvnbaEUe5GPeHLLwoEl37wdJZCEdIQ
IwMXRm54d3gxjS38UsO4CFat0jXzeDYrrM591uxFVOe1U6I/RzQhLU2jtN68nd3OCtnpB/ZkyR1Y
n/E+GnDxyN3z1rEm62SrI1t44PoJYm/QDU/V82oPHJwXFsU9/3YipmRQvSyobz5RnXJuJxlQd2Lx
fEbnE5USoxAeO+HE4Jsy+ZTlw2ctVlwNGQ3ztALgsLtpWUEIw9ZLgYlx2qYsUVNGukgpedBhgMvQ
VP9YB51kld4GMt913otcMg99mbSj8JUYbtC6O3o3i9qg98unFqEFDlOrnogTlb5ymn9J00xv2qp7
9d2naSlRIkiIIiSwmprfjXT0kHUEva/hmZHtXBuOwQGbN7Z9PJor2LdV9C/7ah+xspNRpc/5lrdf
kF6HRUL7DlKQf3p1icoesdP/b2owW1NorvOW4Nbpx9/uQ0aO6ghPxnopFALkyBt78oqnSNS3Wk/+
+6ChZqqnmOJEhiMWo+3vEiJ30xFsap4CDdgVbWytjGQiw1INKRqItWfwWTEnmyVou3hHWMurOQ==
';


$sponge = SHA3::init (SHA3::SHA3_512);
$sponge->absorb (base64_decode (DATA));
$hash = $sponge->squeeze ();
echo bin2hex ($hash), PHP_EOL;

