<?php declare(strict_types=1);

// https://www.khronos.org/registry/typedarray/specs/latest/#6
/**
 * @property-read ArrayBuffer buffer
 * @property-read int byteOffset
 * @property-read int byteLength
 */
abstract class ArrayBufferView
{
    private /* ArrayBuffer */
        $buffer;
    private /* int */
        $byteOffset;
    private /* int */
        $byteLength;

    public function __get(string $propertyName)
    {
        if ($propertyName === "buffer") {
            return $this->buffer;
        } else if ($propertyName === "byteOffset") {
            return $this->byteOffset;
        } else if ($propertyName === "byteLength") {
            return $this->byteLength;
        } else {
            throw new \Exception(self::class . " has no such property '$propertyName'");
        }
    }
}
