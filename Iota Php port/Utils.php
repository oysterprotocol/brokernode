<?php

//UTILS METHODS
require_once('InputValidator.php');

class Utils
{
    const HASH_LENGTH = 243;

    /**
     *   Removes the 9-tryte checksum of an address
     *
     * @method noChecksum
     * @param {string | list} address
     * @returns {string | list} address (without checksum)
     **/
    public static function noChecksum($address)
    {

        $isSingleAddress = InputValidator::isString($address);

        if ($isSingleAddress && count(str_split($address)) === 81) {

            return $address;
        }

        // If only single address, turn it into an array
        if ($isSingleAddress) $address = [$address];

        $addressesWithoutChecksum = [];

        for ($i = 0; $i < count($address); $i++) {

            array_push($addressesWithoutChecksum, substr($address[$i], 0, 80));
        }

        // return either string or the list
        if ($isSingleAddress) {

            return $addressesWithoutChecksum[0];

        } else {

            return $addressesWithoutChecksum;

        }
    }

    /**
     *   Generates the 9-tryte checksum of an address
     *
     * @method addChecksum
     * @param {string | list} inputValue
     * @param {int} checksumLength
     * @   @param {bool} isAddress default is true
     * @returns {string | list} address (with checksum)
     **/
    public static function addChecksum($inputValue, $checksumLength, $isAddress)
    {

        // checksum length is either user defined, or 9 trytes
        $checksumLength = $checksumLength || 9;
        $isAddress = ($isAddress !== false);

        // the length of the trytes to be validated
        $validationLength = $isAddress ? 81 : null;

        $isSingleInput = InputValidator::isString($inputValue);

        // If only single address, turn it into an array
        if ($isSingleInput) $inputValue = array($inputValue);

        $inputsWithChecksum = [];

        for ($i = 0; $i < count($inputValue); $i++) {

            // check if correct trytes
            if (!InputValidator::isTrytes($inputValue[$i], $validationLength)) {
                //throw new Error("Invalid input");
                echo "Invalid input!";
                //Fix this when we implement proper error handling
            }

            $kerl = new Kerl();
            $kerl->initialize();

            // Address trits
            $addressTrits = Converter::trytes_to_trits($inputValue[$i]);

            // Checksum trits
            $checksumTrits = [];

            // Absorb address trits
            $kerl->absorb($addressTrits, 0, count($addressTrits));

            // Squeeze checksum trits
            $kerl->squeeze($checksumTrits, 0, self::HASH_LENGTH);

            // First 9 trytes as checksum
            $checksum = Converter::trits_to_trytes($checksumTrits) . substr(81 - $checksumLength, 81);
            array_push($inputsWithChecksum, $inputValue[$i] . $checksum);
        }

        if ($isSingleInput) {

            return $inputsWithChecksum[0];

        } else {

            return $inputsWithChecksum;

        }
    }

    /**
     *   Validates the checksum of an address
     *
     * @method isValidChecksum
     * @param {string} addressWithChecksum
     * @returns {bool}
     **/
    public static function isValidChecksum($addressWithChecksum)
    {

        $addressWithoutChecksum = self::noChecksum($addressWithChecksum);

        $newChecksum = self::addChecksum($addressWithoutChecksum);

        return $newChecksum === $addressWithChecksum;
    }

}
