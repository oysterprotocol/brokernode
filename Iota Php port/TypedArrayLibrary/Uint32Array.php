<?php declare(strict_types=1);

require_once('TypedArray.php');

// https://www.khronos.org/registry/typedarray/specs/latest/#7
class Uint32Array extends TypedArray
{
    const BYTES_PER_ELEMENT = 4;
    const ELEMENT_PACK_CODE = 'L';
}
