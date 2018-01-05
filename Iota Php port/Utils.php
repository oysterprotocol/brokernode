<?php


require_once('InputValidator.php');
require_once('kerl-php/kerl.php');

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
        echo "<br/>signatureMessageFragment is <br/>" . $transaction->signatureMessageFragment . "<br/>";
        echo "<br/>address is <br/>" . $transaction->address . "<br/>";
        echo "<br/>trits_to_trytes(valueTrits) is <br/>" . (Converter::trits_to_trytes($valueTrits)) . "<br/>";
        echo "<br/>transaction->obsoleteTag is <br/>" . $transaction->obsoleteTag . "<br/>";
        echo "<br/>trits_to_trytes(timestampTrits) is <br/>" . Converter::trits_to_trytes($timestampTrits) . "<br/>";
        echo "<br/>trits_to_trytes(currentIndexTrits) is <br/>" . Converter::trits_to_trytes($currentIndexTrits) . "<br/>";
        echo "<br/>trits_to_trytes(lastIndexTrits) is <br/>" . Converter::trits_to_trytes($lastIndexTrits) . "<br/>";
        echo "<br/>transaction->bundle is <br/>" . $transaction->bundle . "<br/>";
        echo "<br/>transaction->trunkTransaction is <br/>" . $transaction->trunkTransaction . "<br/>";
        echo "<br/>transaction->branchTransaction is <br/>" . $transaction->branchTransaction . "<br/>";
        echo "<br/>transaction->tag is <br/>" . $transaction->tag . "<br/>";
        echo "<br/>trits_to_trytes(attachmentTimestampTrits) is <br/>" . Converter::trits_to_trytes($attachmentTimestampTrits) . "<br/>";
        echo "<br/>trits_to_trytes(attachmentTimestampLowerBoundTrits) is <br/>" . Converter::trits_to_trytes($attachmentTimestampLowerBoundTrits) . "<br/>";
        echo "<br/>trits_to_trytes(attachmentTimestampUpperBoundTrits) is <br/>" . Converter::trits_to_trytes($attachmentTimestampUpperBoundTrits) . "<br/>";
        echo "<br/>transaction->nonce is <br/>" . $transaction->nonce . "<br/>";
        */

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
}
