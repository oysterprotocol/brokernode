<?php declare(strict_types=1);

require_once('TypedArray.php');

class Uint8Array extends TypedArray
{
    const BYTES_PER_ELEMENT = 1;
    const ELEMENT_PACK_CODE = 'C';
}