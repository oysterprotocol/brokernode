<?php


require_once('InputValidator.php');
require_once('Converter.php');
require_once('kerl-php/kerl.php');
require_once('kerl-php/conv.php');

//UTILS METHODS
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

        if ($isSingleAddress && strlen($address) === 81) {

            return $address;
        }

        // If only single address, turn it into an array
        if ($isSingleAddress) $address = [$address];

        $addressesWithoutChecksum = [];

        for ($i = 0; $i < count($address); $i++) {

            array_push($addressesWithoutChecksum, substr($address[$i], 0, 81));
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
     * @param {bool} isAddress default is true
     * @returns {string | list} address (with checksum)
     **/
    public static function addChecksum($inputValue, $checksumLength = 9, $isAddress = true)
    {
        // checksum length is either user defined, or 9 trytes
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
                echo "EXCEPTION -- Invalid input!";
                //Fix this when we implement proper error handling
            }

            $kerl = new Kerl();

            // Address trits
            $addressTrits = Converter::trytes_to_trits($inputValue[$i]);

            // Checksum trits
            $checksumTrits = [];

            // Absorb address trits
            $kerl->absorb($addressTrits, 0, count($addressTrits));

            // Squeeze checksum trits
            $kerl->squeeze($checksumTrits, 0, self::HASH_LENGTH);

            // First 9 trytes as checksum
            $checksum = substr(trits_to_trytes($checksumTrits), 81 - $checksumLength, 81);

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

    /**
     *   Converts a transaction object into trytes
     *
     * @method transactionTrytes
     * @param {object} transactionTrytes
     * @returns {String} trytes
     **/
    public static function transactionTrytes($transaction)
    {
        $valueTrits = Converter::trytes_to_trits($transaction->value);
        while (is_null($valueTrits) || count($valueTrits) < 81) {
            $valueTrits[] = 0;
        }

        $timestampTrits = Converter::trytes_to_trits($transaction->timestamp);
        while (is_null($timestampTrits) || count($timestampTrits) < 27) {
            $timestampTrits[] = 0;
        }

        $currentIndexTrits = Converter::trytes_to_trits($transaction->currentIndex);
        while (is_null($currentIndexTrits) || count($currentIndexTrits) < 27) {
            $currentIndexTrits[] = 0;
        }

        $lastIndexTrits = Converter::trytes_to_trits($transaction->lastIndex);
        while (is_null($lastIndexTrits) || count($lastIndexTrits) < 27) {
            $lastIndexTrits[] = 0;
        }

        $attachmentTimestampTrits = Converter::trytes_to_trits($transaction->attachmentTimestamp || 0);
        while (is_null($attachmentTimestampTrits) || count($attachmentTimestampTrits) < 27) {
            $attachmentTimestampTrits[] = 0;
        }

        $attachmentTimestampLowerBoundTrits = Converter::trytes_to_trits($transaction->attachmentTimestampLowerBound || 0);
        while (is_null($attachmentTimestampLowerBoundTrits) || count($attachmentTimestampLowerBoundTrits) < 27) {
            $attachmentTimestampLowerBoundTrits[] = 0;
        }

        $attachmentTimestampUpperBoundTrits = Converter::trytes_to_trits($transaction->attachmentTimestampUpperBound || 0);
        while (is_null($attachmentTimestampUpperBoundTrits) || count($attachmentTimestampUpperBoundTrits) < 27) {
            $attachmentTimestampUpperBoundTrits[] = 0;
        }

        $transaction->tag = $transaction->tag ? $transaction->tag : $transaction->obsoleteTag;

        /*
         * Leaving this in for future debugging purposes, makes it easier to find where problems are
         *
         */

//        $nl =  "<br/>";
//
//        echo $nl . "signatureMessageFragment is " . $nl . $transaction->signatureMessageFragment .  $nl;
//        echo $nl . "address is " . $nl . $transaction->address .  $nl;
//        echo $nl . "trits_to_trytes(valueTrits) is " . $nl . (Converter::trits_to_trytes($valueTrits)) .  $nl;
//        echo $nl . "transaction->obsoleteTag is " . $nl . $transaction->obsoleteTag .  $nl;
//        echo $nl . "trits_to_trytes(timestampTrits) is " . $nl . Converter::trits_to_trytes($timestampTrits) .  $nl;
//        echo $nl . "trits_to_trytes(currentIndexTrits) is " . $nl . Converter::trits_to_trytes($currentIndexTrits) .  $nl;
//        echo $nl . "trits_to_trytes(lastIndexTrits) is " . $nl . Converter::trits_to_trytes($lastIndexTrits) .  $nl;
//        echo $nl . "transaction->bundle is " . $nl . $transaction->bundle .  $nl;
//        echo $nl . "transaction->trunkTransaction is " . $nl . $transaction->trunkTransaction .  $nl;
//        echo $nl . "transaction->branchTransaction is " . $nl . $transaction->branchTransaction .  $nl;
//        echo $nl . "transaction->tag is " . $nl . $transaction->tag .  $nl;
//        echo $nl . "trits_to_trytes(attachmentTimestampTrits) is " . $nl . Converter::trits_to_trytes($attachmentTimestampTrits) .  $nl;
//        echo $nl . "trits_to_trytes(attachmentTimestampLowerBoundTrits) is " . $nl . Converter::trits_to_trytes($attachmentTimestampLowerBoundTrits) .  $nl;
//        echo $nl . "trits_to_trytes(attachmentTimestampUpperBoundTrits) is " . $nl . Converter::trits_to_trytes($attachmentTimestampUpperBoundTrits) .  $nl;
//        echo $nl . "transaction->nonce is " . $nl . $transaction->nonce .  $nl;


        return $transaction->signatureMessageFragment
            . $transaction->address
            . Converter::trits_to_trytes($valueTrits)
            . $transaction->obsoleteTag
            . Converter::trits_to_trytes($timestampTrits)
            . Converter::trits_to_trytes($currentIndexTrits)
            . Converter::trits_to_trytes($lastIndexTrits)
            . $transaction->bundle
            . $transaction->trunkTransaction
            . $transaction->branchTransaction
            . $transaction->tag
            . Converter::trits_to_trytes($attachmentTimestampTrits)
            . Converter::trits_to_trytes($attachmentTimestampLowerBoundTrits)
            . Converter::trits_to_trytes($attachmentTimestampUpperBoundTrits)
            . $transaction->nonce;
    }

    /**
     *   Converts transaction trytes of 2673 trytes into a transaction object
     *
     * @method transactionObject
     * @param {string} trytes
     * @returns {String} transactionObject
     **/
    public static function transactionObject($trytes)
    {

        if (!$trytes) return;

        // validity check
        for ($i = 2279; $i < 2295; $i++) {
            if ($trytes{$i} !== "9") {
                return null;
            }
        }

        $thisTransaction = new stdClass();

        /*
         * NOTICE:  all of the code that has been commented out in this file
         * produces the wrong output.  I've left it in the file so that it's closer to
         * being a 1-to-1 port of the iota.lib.js library's Utils.transactionObject method.
         * If we need the other properties and if we need them to be correct, we need to
         * revisit this method.
         */

//        $transactionTrits = Converter::trytes_to_trits($trytes);
//        $hash = array();
//
//        $curl = new Kerl();  // This needs to be the "Curl" library that Iota JS uses
//
//        // generate the correct transaction hash
//        $curl->absorb($transactionTrits, 0, count($transactionTrits));
//        $curl->squeeze($hash, 0, 243);
//
//        echo implode(", ", $hash);
//
//        $thisTransaction->hash = trits_to_trytes($hash);
        $thisTransaction->signatureMessageFragment = substr($trytes, 0, 2187 - 0);
        $thisTransaction->address = substr($trytes, 2187, 2268 - 2187);
        //$thisTransaction->value = Converter::trits_to_integers(array_slice($transactionTrits, 6804, count($transactionTrits) - 6804));
        $thisTransaction->obsoleteTag = substr($trytes, 2295, 2322 - 2295);
        //$thisTransaction->timestamp = Converter::trits_to_integers(array_slice($transactionTrits, 6966, count($transactionTrits) - 6966));
        //$thisTransaction->currentIndex = Converter::trits_to_integers(array_slice($transactionTrits, 6993, count($transactionTrits) - 6993));
        //$thisTransaction->lastIndex = Converter::trits_to_integers(array_slice($transactionTrits, 7020, count($transactionTrits) - 7020));
        $thisTransaction->bundle = substr($trytes, 2349, 2430 - 2349);
        $thisTransaction->trunkTransaction = substr($trytes, 2430, 2511 - 2430);
        $thisTransaction->branchTransaction = substr($trytes, 2511, 2592 - 2511);
        $thisTransaction->tag = substr($trytes, 2592, 2619 - 2592);
        //$thisTransaction->attachmentTimestamp = Converter::trits_to_integers(array_slice($transactionTrits, 7857, count($transactionTrits) - 7857));
        //$thisTransaction->attachmentTimestampLowerBound = Converter::trits_to_integers(array_slice($transactionTrits, 7884, count($transactionTrits) - 7884));
        //$thisTransaction->attachmentTimestampUpperBound = Converter::trits_to_integers(array_slice($transactionTrits, 7911, count($transactionTrits) - 7911));
        $thisTransaction->nonce = substr($trytes, 2646, 2673 - 2646);

        return $thisTransaction;
    }
}
