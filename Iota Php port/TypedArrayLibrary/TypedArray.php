<?php declare(strict_types=1);

require_once("ArrayBufferView.php");

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
            throw new \InvalidArgumentException("Integer, " . TypedArray::class . " or " . ArrayBuffer::clsss . " expected for first parameter, " . gettype($array) . " given");
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
