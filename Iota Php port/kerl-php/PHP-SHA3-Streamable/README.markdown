# Streamable SHA-3 in pure PHP!

* **Native 64-bit acceleration available**
* Available as a namespaced simple PHP class licensed under **LGPLv3 or later**
* Supported: PHP 5.2 (Warning: EOL in January 2011) or later (legacy)
* Namespaced class: PHP 5.3 or later
* Acceleration supported: PHP 5.6.3 or later (64 bit)
* Best performance with PHP 7.0 or later
* The fastest known implementation in pure PHP: faster than `https://github.com/0xbb/php-sha3` (PHP 5.6.3 or PHP 7.0)


## Features

* SHA3-224
* SHA3-256 (for 128-bit security like SHA2-256)
* SHA3-384
* SHA3-512 (for 256-bit security like SHA2-512)
* SHAKE128 (Unbounded output; use at least 256 bits (32 bytes) of output for 128-bit security)
* SHAKE256 (Unbounded output; use at least 512 bits (64 bytes) of output for 256-bit security)


## Performance

Benchmarking result:
`PHP 7.0 (64 bit) > PHP 7.0 (32 bit) > PHP 5.6.3 (64 bit) >> PHP 5.6 (32 bit) > PHP 5.2`

**NEW: Use 64-bit PHP for best performance, because native 64-bit integers will be used. (Only supported on PHP 5.6.3 or later)** 64-bit mode is more than 4 times faster in PHP 5.6.

Note that SHA-3 is considably slower than SHA-2 in *software*, not to mention
PHP's notable slowness. PHP is only fast when your program is trivial or I/O-bound.


## Usage
Namespaced class `desktopd\SHA3\Sponge` is available at the `namespaced/` directory of this repository. **This is the optimized version.**

        <?php
        use desktopd\SHA3\Sponge as SHA3;
        require __DIR__ . '/namespaced/desktopd/SHA3/Sponge.php';
        
        $sponge = SHA3::init (SHA3::SHA3_512);
        $sponge->absorb ('');
        // fixed size (512 bits) output
        echo bin2hex ($sponge->squeeze ()), PHP_EOL;
        // a69f73cca23a9ac5c8b567dc185a756e97c982164fe25859e0d1dcc1475c80a615b2123af1f5f94c11e3e9402c3ac558f500199d95b6d3e301758586281dcd26

### Documentation
`static public desktopd\SHA3\Sponge::init ($type) : desktopd\SHA3\Sponge`
where `$type` is one of:

* desktopd\SHA3\Sponge::SHA3_224
Output: 224 bits (28 bytes)
* desktopd\SHA3\Sponge::SHA3_256
Output: 256 bits (32 bytes)
* desktopd\SHA3\Sponge::SHA3_384
Output: 384 bits (48 bytes)
* desktopd\SHA3\Sponge::SHA3_512
Output: 512 bits (64 bytes)
* desktopd\SHA3\Sponge::SHAKE128
Output: As much as you want (speicify the length in bytes with `squeeze()`)
* desktopd\SHA3\Sponge::SHAKE256
Output: As much as you want (speicify the length in bytes with `squeeze()`)


`public desktopd\SHA3\Sponge::absorb (string $data) : desktopd\SHA3\Sponge`
Add data to hash. You can call this as many times as you need **before calling `squeeze ()`**.


`public desktopd\SHA3\Sponge::squeeze ([int $length]) : string`
Get hash data. Only specify `$length` with SHAKE128 or SHAKE256. With SHAKE128 or SHAKE256, you can call this as many times as you need.


### PHP 5.2 (legacy)
Use `SHA3.php` **(slower)**:

        <?php
        require dirname (__FILE__) . '/SHA3.php';
        
        $shake128 = SHA3::init (SHA3::SHAKE128);
        $shake128->absorb ('The quick brown fox ');
        $shake128->absorb ('jumps over the lazy dog');
        // Squeeze 256 bits of hash output!
        echo bin2hex ($shake128->squeeze (32)) . PHP_EOL;
        // f4202e3c5852f9182a0430fd8144f0a74b95e7417ecae17db0f8cfeed0e3e66e
        // more output possible


## FAQ
### Why so slow?
This is nearly the fastest in pure PHP. (Without depending on an extension)
Long answer: PHP is not good at processing large binary data.

### Why do you use the words 'sponge', 'absorb', or 'squeeze'?
The terms stem from the fact that SHA-3 is based on *sponge construction*.
Read more [here](http://sponge.noekeon.org/).


## Support
Please [report](https://notabug.org/desktopd/PHP-SHA3-Streamable/issues) if you find a problem.


## TODO
* Optimize even more for speed, while defending against timing attacks (SHA-3 is good at that)


## Legal

README.markdown:
Copyright (C) 2016 Desktopd Developers

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.


PHP-SHA3-Streamable is free software. The library (the files `SHA3.php` and
`namespaced/desktopd/SHA3/Sponge.php`) is under GNU LGPL version 3 or later.
Other codes are under GNU GPL version 3 or later.
See the files COPYING.LGPL (GNU LGPL) and COPYING (GNU GPL) for copying
conditions.
