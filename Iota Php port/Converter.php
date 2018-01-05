<?php

class Converter
{
    const RADIX = 3;
    const RADIX_BYTES = 256;
    const MAX_TRIT_VALUE = 1;
    const MIN_TRIT_VALUE = -1;
    const BYTE_HASH_LENGTH = 48;
    const TRYTES_ALPHABET = '9ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    const TRYTES_TRITS = [
        [0, 0, 0],
        [1, 0, 0],
        [-1, 1, 0],
        [0, 1, 0],
        [1, 1, 0],
        [-1, -1, 1],
        [0, -1, 1],
        [1, -1, 1],
        [-1, 0, 1],
        [0, 0, 1],
        [1, 0, 1],
        [-1, 1, 1],
        [0, 1, 1],
        [1, 1, 1],
        [-1, -1, -1],
        [0, -1, -1],
        [1, -1, -1],
        [-1, 0, -1],
        [0, 0, -1],
        [1, 0, -1],
        [-1, 1, -1],
        [0, 1, -1],
        [1, 1, -1],
        [-1, -1, 0],
        [0, -1, 0],
        [1, -1, 0],
        [-1, 0, 0]
    ];

    public function __construct()
    {
    }

    public static function trits_to_integers($trits_param)
    {

        $returnValue = 0;
        $three = 3;

        for ($i = count($trits_param) - 1; $i >= 0; --$i) {

            $returnValue = $returnValue * $three + $trits_param[$i];
        }

        return $returnValue;
    }

    public static function integer_to_trits($value)
    {

        $destination = [];
        $absoluteValue = $value < 0 ? $value * -1 : $value;
        $i = 0;

        while ($absoluteValue > 0) {

            $remainder = ($absoluteValue % self::RADIX);
            $absoluteValue = floor($absoluteValue / self::RADIX);

            if ($remainder > self::MAX_TRIT_VALUE) {

                $remainder = self::MIN_TRIT_VALUE;
                $absoluteValue++;

            }

            $destination[$i] = $remainder;
            $i++;
        }

        if ($value < 0) {

            for ($j = 0; $j < count($destination); $j++) {

                // switch values
                $destination[$j] = $destination[$j] === 0 ? 0 : -$destination[$j];
            }
        }

        return $destination;
    }

    public static function trits_to_trytes($trits_param)
    {

        $trytes = '';

        $TRYTES_ALPHABET_ARRAY = str_split(self::TRYTES_ALPHABET);

        for ($i = 0, $lengthParam = count($trits_param); $i < $lengthParam; $i = $i + 3) {

            for ($j = 0,
                 $lengthAlphabetArray = count($TRYTES_ALPHABET_ARRAY);
                 $j < $lengthAlphabetArray; ++$j) {

                if (self::TRYTES_TRITS[$j][0] === $trits_param[$i] &&
                    self::TRYTES_TRITS[$j][1] === $trits_param[$i + 1] &&
                    self::TRYTES_TRITS[$j][2] === $trits_param[$i + 2]) {
                    $trytes = $trytes . self::TRYTES_ALPHABET[$j];
                    break;

                }
            }
        }

        return $trytes;
    }

    public static function trytes_to_trits($trytes_param, $state = [])
    {

        $trits = $state || [];

        if (is_int($trytes_param)) {

            $absoluteValue = $trytes_param < 0 ? -$trytes_param : $trytes_param;

            while ($absoluteValue > 0) {

                $remainder = $absoluteValue % 3;
                $absoluteValue = floor($absoluteValue / 3);

                if ($remainder > 1) {
                    $remainder = -1;
                    $absoluteValue++;
                }

                $trits[] = $remainder;
            }
            if ($trytes_param < 0) {

                for ($i = 0; $i < count($trits); $i++) {

                    $trits[$i] = -$trits[$i];
                }
            }
        } else {

            for ($i = 0; $i < strlen($trytes_param); $i++) {

                $index = strpos(self::TRYTES_ALPHABET, substr($trytes_param, $i, 1));

                $trits[$i * 3] = self::TRYTES_TRITS[$index][0];
                $trits[$i * 3 + 1] = self::TRYTES_TRITS[$index][1];
                $trits[$i * 3 + 2] = self::TRYTES_TRITS[$index][2];

            }

        }
        return $trits;
    }
}


