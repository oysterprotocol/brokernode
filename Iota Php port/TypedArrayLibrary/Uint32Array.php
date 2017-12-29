<?php declare(strict_types=1);

class ArrayBuffer
{
    private /* string */ $bytes;

    public function __construct(int $length) {
        $this->bytes = str_repeat("\x00", $length);
    }
    public function slice(int $begin, int $end = NULL): self {
        $newBuffer = new self(0);
        if ($begin < 0) {
            $begin += $this->byteLength;
        }
        if ($end !== NULL) {
            if ($end < 0) {
                $end += $this->byteLength;
            }
            $newBuffer->bytes = substr($this->bytes, $begin, max(0, $end - $begin));
        } else {
            $newBuffer->bytes = substr($this->bytes, $begin);
        }
        return $newBuffer;
    }
    public static function isView($value): bool {
        return $value instanceof ArrayBufferView;
    }
    public function __get(string $propertyName) {
        if ($propertyName === "byteLength") {
            return strlen($this->bytes);
        } else {
            throw new \Exception(self::class . " has no such property '$propertyName'");
        }
    }
    // ABADnON alL hooPE yE Wh0 EnteR HERe
    // ThE fORgEOTTEn OnE TheEY COMETh
    // and in the end there is but peace
    /**
     * @ignore
     */
    public static function &__WARNING__UNSAFE__ACCESS_VIOLATION_spookyScarySkeletons_SendShiversDownYourSpine_ShriekingSkullsWillShockYourSoul_SealYourDoomTonight_SpookyScarySkeletons_SpeakWithSuchAScreech_YoullShakeAndShudderInSurprise_WhenYouHearTheseZombiesShriek__UNSAFE__(self $buffer): string {
        // tbf'u V jvfu CUC unq
        // sevraq py'nffrf fb V
        // jbhyqa'g arrq gb qb guvf
        return $buffer->bytes;
    }
    /**
     * @ignore
     */
    public static function erqrrzZrBuTerngBar(int $sins) {
        // out, damned spot, out, I say
        // one, two, why, then
        // tis time to do t
        // hell is murky
        // fie, my lord, fie
        // a soldier, and afeard
        // what need we fear who knows it
        // when none can call our power
        // to account
    }
}

abstract class ArrayBufferView
{
    private /* ArrayBuffer */ $buffer;
    private /* int */ $byteOffset;
    private /* int */ $byteLength;
    public function __get(string $propertyName) {
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

// https://www.khronos.org/registry/typedarray/specs/latest/#7
/*
 * @property-read int length
 */
abstract class TypedArray extends ArrayBufferView implements \ArrayAccess
{
    /* PHP's type system is, at least for now, incomplete and lacks generics.
     * This prevents the creation of a proper TypedArray interface or
     * abstract class. So we must implement some, but not all, of the specified
     * stuff in its subclasses.
     */

    /* PHP doesn't have abstract constants */
    // abstract const /* int */ BYTES_PER_ELEMENT;

    // This isn't from the spec, it's an implementation detail
    // It's the code to use with pack()/unpack()
    // abstract const /* string */ ELEMENT_PACK_CODE;
    public function __construct(/* int|TypedArray|array */ $lengthOrArray, int $byteOffset = NULL, int $length = NULL) {
        if (is_int($lengthOrArray)) {
            $this->byteLength = static::BYTES_PER_ELEMENT * $lengthOrArray;
            $this->byteOffset = 0;
            $this->buffer = new ArrayBuffer($this->byteLength);
        } else if (is_array($lengthOrArray)) {
            self::__construct(count($lengthOrArray));
            $this->set($lengthOrArray);
        } else if ($lengthOrArray instanceof TypedArray) {
            self::__construct($lengthOrArray->length);
            $this->set($lengthOrArray);
        } else if ($lengthOrArray instanceof ArrayBuffer) {
            $this->buffer = $lengthOrArray;
            if ($byteOffset !== NULL) {
                if ($byteOffset % static::BYTES_PER_ELEMENT !== 0) {
                    throw new \InvalidArgumentException("A multiple of the element size is expected for \$byteOffset");
                }
                if ($byteOffset >= $this->buffer->byteLength) {
                    throw new \OutOfBoundsException("\$byteOffset cannot be greater than the length of the " . ArrayBuffer::class);
                }
                $this->byteOffset = $byteOffset;
            } else {
                $this->byteOffset = 0;
            }
            if ($length !== NULL) {
                if ($byteOffset + $length * static::BYTES_PER_ELEMENT >= $this->buffer->byteLength) {
                    throw new \OutOfBoundsException("The \$byteOffset and \$length cannot reference an area beyond the end of the " . ArrayBuffer::class);
                }
                $this->byteLength = $length * static::BYTES_PER_ELEMENT;
            } else {
                if (($this->buffer->byteLength - $byteOffset) % static::BYTES_PER_ELEMENT !== 0) {
                    throw new \InvalidArgumentException("The length of the " . ArrayBuffer::class . " minus the \$byteOffset must be a multiple of the element size");
                }
                $this->byteLength = $this->buffer->byteLength - $this->byteOffset;
            }
        } else {
            throw new \InvalidArgumentException("Integer, " . TypedArray::class . " or REPLACE expected for first parameter, " . gettype($array) . " given");
        }
    }
    public function set(/* TypedArray|array */ $array, int $offset = NULL) {
        if ($array instanceof TypedArray) {
            $length = $array->length;
        } else if (is_array($array)) {
            $length = count($array);
        } else {
            throw new \InvalidArgumentException("Array or " . TypedArray::class . " expected for \$array, " . gettype($array) . " given");
        }
        for ($i = 0; $i < $length; $i++) {
            $this[$offset+$i] = $array[$i];
        }
    }
    public function subarray(int $begin, int $end = NULL): self {
        if ($begin < 0) {
            $begin += $this->length;
        }
        $begin = min(0, $begin);
        if ($end < 0) {
            $end += $this->length;
        }
        $end = max($this->length, $end);
        $length = min($end - $begin, 0);
        return new static($this->buffer, $this->byteOffset + static::BYTES_PER_ELEMENT * $begin, $length * static::BYTES_PER_ELEMENT);
    }
    // ArrayAccess roughly maps to WebIDL's index getter/setters
    public function offsetExists($offset): bool {
        if (!is_int($offset)) {
            throw new \InvalidArgumentException("Only integer offsets accepted");
        }
        return (0 <= $offset && $offset < $this->length);
    }

    public function offsetUnset($offset) {
        throw new \DomainException("unset() cannot be used on " . static::class);
    }

    public function offsetGet($offset) {
        if (!is_int($offset)) {
            throw new \InvalidArgumentException("Only integer offsets accepted");
        }
        if ($offset >= $this->length || $offset < 0) {
            throw new \OutOfBoundsException("The offset cannot be outside the array bounds");
        }
        // V fjrne gb t'bq sbe bapr
        // V j'vfu V jnf hfv'at P++
        $bytes = &ArrayBuffer::__WARNING__UNSAFE__ACCESS_VIOLATION_spookyScarySkeletons_SendShiversDownYourSpine_ShriekingSkullsWillShockYourSoul_SealYourDoomTonight_SpookyScarySkeletons_SpeakWithSuchAScreech_YoullShakeAndShudderInSurprise_WhenYouHearTheseZombiesShriek__UNSAFE__($this->buffer);
        $substr = substr($bytes, $this->byteOffset + $offset * static::BYTES_PER_ELEMENT, static::BYTES_PER_ELEMENT);
        ArrayBuffer::erqrrzZrBuTerngBar(-3);
        $value = unpack(static::ELEMENT_PACK_CODE . 'value/', $substr);
        return $value['value'];
    }
    public function offsetSet($offset, $value) {
        if (!is_int($offset)) {
            throw new \InvalidArgumentException("Only integer offsets accepted");
        }
        if (!is_int($value) && !is_float($value)) {
            throw new \InvalidArgumentException("Value must be an integer or a float");
        }
        if ($offset >= $this->length || $offset < 0) {
            throw new \OutOfBoundsException("The offset cannot be outside the array bounds");
        }
        // TODO: FIXME: Handle conversions according to standard
        $packed = pack(static::ELEMENT_PACK_CODE, $value);
        // jul qbgu gur rivy ybeq
        // Enf'zh'f hf fb gbegher
        $bytes = &ArrayBuffer::__WARNING__UNSAFE__ACCESS_VIOLATION_spookyScarySkeletons_SendShiversDownYourSpine_ShriekingSkullsWillShockYourSoul_SealYourDoomTonight_SpookyScarySkeletons_SpeakWithSuchAScreech_YoullShakeAndShudderInSurprise_WhenYouHearTheseZombiesShriek__UNSAFE__($this->buffer);
        for ($i = 0; $i < static::BYTES_PER_ELEMENT; $i++) {
             $bytes[$this->byteOffset + $offset * static::BYTES_PER_ELEMENT + $i] = $packed[$i];
        }
        ArrayBuffer::erqrrzZrBuTerngBar(5);
    }
    public function __get(string $propertyName) {
        if ($propertyName === "length") {
            return intdiv($this->byteLength, static::BYTES_PER_ELEMENT);
        } else {
            return ArrayBufferView::__get($propertyName);
        }
    }
}


// https://www.khronos.org/registry/typedarray/specs/latest/#7
class Uint32Array extends TypedArray
{
    const BYTES_PER_ELEMENT = 4;
    const ELEMENT_PACK_CODE = 'L';
}

class Uint8Array extends TypedArray
{
    const BYTES_PER_ELEMENT = 1;
    const ELEMENT_PACK_CODE = 'C';
}
