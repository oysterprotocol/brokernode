<?php
//InputValidator METHODS

class InputValidator
{

    public function _constructor()
    {
    }

    /*Not all iota library input validator methods are in this file.
    Just the ones needed for prepareTransfers.

    /**
    *   checks if input is correct hash
    *
    *   @method isHash
    *   @param {string} hash
    *   @returns {boolean}
    **/
    public static function isHash($hash)
    {

        // Check if valid, 81 trytes
        if (!self::isTrytes($hash, 81)) {

            return false;
        }

        return true;
    }

    /**
     *   checks if input is list of correct trytes
     *
     * @method isArrayOfHashes
     * @param {list} hashesArray
     * @returns {boolean}
     **/
    public static function isArrayOfHashes($hashesArray)
    {

        if (!is_array($hashesArray)) {
            return false;
        }

        for ($i = 0; $i < $hashesArray->length; $i++) {

            $hash = $hashesArray[$i];

            // Check if address with checksum
            if (count(str_split($hash)) === 90) {

                if (!self::isTrytes($hash, 90)) {
                    return false;
                }
            } else {

                if (!self::isTrytes($hash, 81)) {
                    return false;
                }
            }
        }

        return true;
    }


    /**
     *   checks if correct inputs list
     *
     * @method isInputs
     * @param {array} inputs
     * @returns {boolean}
     **/
    public static function isInputs($inputs)
    {

        if (!is_array($inputs)) {
            return false;
        }

        for ($i = 0; $i < $inputs->length; $i++) {

            $input = $inputs[$i];

            // If input does not have keyIndex and address, return false
            if (!property_exists($input, 'security') || !property_exists($input, 'keyIndex') || !property_exists($input, 'address')) {
                return false;
            }

            if (!self::isAddress($input->address)) {
                return false;
            }

            if (!self::isValue($input->security)) {
                return false;
            }

            if (!self::isValue($input->keyIndex)) {
                return false;
            }
        }

        return true;
    }


    /**
     *   checks if input is list of correct trytes
     *
     * @method isArrayOfTrytes
     * @param {list} trytesArray
     * @returns {boolean}
     **/
    public static function isArrayOfTrytes($trytesArray)
    {

        if (!is_array($trytesArray)) {
            return false;
        }

        for ($i = 0; $i < $trytesArray->length; $i++) {

            $tryteValue = $trytesArray[$i];

            // Check if correct 2673 trytes
            if (!self::isTrytes($tryteValue, 2673)) {
                return false;
            }
        }

        return true;
    }


    /**
     *   checks if integer value
     *
     * @method isValue
     * @param {string} value
     * @returns {boolean}
     **/
    public static function isValue($value)
    {

        // check if correct number
        return is_int($value);
    }


    /**
     *   checks if input is correct hash
     *
     * @method isTransfersArray
     * @param {array} hash
     * @returns {boolean}
     **/
    public static function isTransfersArray($transfersArray)
    {

        if (!is_array($transfersArray)) {
            return false;
        }

        for ($i = 0; $i < $transfersArray->length; $i++) {

            $transfer = $transfersArray[$i];

            // Check if valid address
            $address = $transfer->address;
            if (!self::isAddress($address)) {
                return false;
            }

            // Validity check for value
            $value = $transfer->value;
            if (!self::isValue($value)) {
                return false;
            }

            // Check if message is correct trytes of any length
            $message = $transfer->message;
            if (!self::isTrytes($message, "0,")) {
                return false;
            }

            // Check if tag is correct trytes of {0,27} trytes
            $tag = $transfer->tag || $transfer->obsoleteTag;
            if (!self::isTrytes($tag, "0,27")) {
                return false;
            }

        }

        return true;
    }


    /**
     *   checks if input is correct address
     *
     * @method isAddress
     * @param {string} address
     * @returns {boolean}
     **/
    public static function isAddress($address)
    {
        // TODO: In the future check checksum

        // Check if address with checksum
        if (count(str_split($address)) === 90) {

            if (!self::isTrytes($address, 90)) {
                return false;
            }
        } else {

            if (!self::isTrytes($address, 81)) {
                return false;
            }
        }

        return true;
    }

    /**
     *   checks if input is correct trytes consisting of A-Z9
     *   optionally validate length
     *
     * @method isTrytes
     * @param {string} trytes
     * @param {integer} length optional
     * @returns {boolean}
     **/
    public static function isTrytes($trytes, $length = "0,")
    { //length is optional and defaults to "0,"

        $regexTrytesPattern = "/^[9A-Z]{" . $length . "}$/";
        return preg_match($regexTrytesPattern, $trytes);
    }

    /**
     *   checks whether input is a string or not
     *
     * @method isString
     * @param {string}
     * @returns {boolean}
     **/
    public static function isString($string)
    {

        return gettype($string) === 'string';
    }

}

//tests of isAddress
$testTrytes1 = 'AAJAAAJAJAAAJAAAJAJAAAJAAAJAJAAJAAAJAJJAAAJAJAAAJAAAJAJAAAJAAAJAJAAAJAAAJAJAAAJAA'; //81
$testTrytes2 = 'AAJARRRRRRRRRAAJAJAAAJAAAJAJAAAJAAAJAJAAJAAAJAJJAAAJAJAAAJAAAJAJAAAJAAAJAJAAAJAAAJAJAAAJAA'; //90
$testTrytes3 = 'AAJARRRRRRRRRAAJAJAAAJAAAJAJAAAJAAAJAJAAJAAAJAJJAAAJAJAAAJAAAJAJAAAJAAAJAJAAAJAAAJAJAAAAJAA'; //91

$validator = new InputValidator();

$result1 = $validator->isAddress($testTrytes1);
$result2 = $validator->isAddress($testTrytes2);
$result3 = $validator->isAddress($testTrytes3);
if ($result1 == 1)
    echo "Test Success!"; //we expected true
else
    echo "Test Failed!"; //we expected true
if ($result2 == 1)
    echo "Test Success!"; //we expected true
else
    echo "Test Failed!"; //we expected true
if ($result3 == 1)
    echo "Test Failed!"; //we expected false
else
    echo "Test Success!"; //we expected false
//end tests


//tests of isTrytes
$testTrytes1 = 'AAJAAAJAJAAAJAAAJAJAAAJAAAJAJAAJAAAJAJJAAAJAJAAAJAAAJAJAAAJAAAJAJAAAJAAAJAJAAAJAA';
$testTrytes2 = 'AAJAAAJAJAAAJAAAJAJAAAJAAAJAJAAJAAAJAJJAAAJAJAAAJAAAJAJAAAJAAAJAJAAAJAAAJAJAAAJAA7'; //has a 7
$result4 = $validator->isTrytes($testTrytes1);
$result5 = $validator->isTrytes($testTrytes2);
if ($result4 == 1)
    echo "Test Success!"; //we expected true
else
    echo "Test Failed!"; //we expected true
if ($result5 == 1)
    echo "Test Failed!"; //we expected false
else
    echo "Test Success!"; //we expected false
//end tests



