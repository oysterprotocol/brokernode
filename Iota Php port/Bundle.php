<?php

require_once('Converter.php');
//require_once('Kerl.php');
require_once('Helper.php');

class Bundle
{

    const HASH_LENGTH = 243;

    public function __construct()
    {
        // Access instance variables with $this
        $this->bundle = [];
        $this->converter = new Converter();
        $this->words = new Words();
    }

    public function addEntry($signatureMessageLength, $address, $value, $tag, $timestamp, $index)
    {

        for ($i = 0; $i < $signatureMessageLength; $i++) {

            $transactionObject = new stdClass();
            $transactionObject->address = $address;
            $transactionObject->value = $i == 0 ? $value : 0;
            $transactionObject->obsoleteTag = $tag;
            $transactionObject->tag = $tag;
            $transactionObject->timestamp = $timestamp;

            $this->bundle[count($this->bundle)] = $transactionObject;
        }
    }

    public function addTrytes($signatureFragments)
    {

        $emptySignatureFragment = '';
        $emptyHash = '999999999999999999999999999999999999999999999999999999999999999999999999999999999';
        $emptyTag = str_pad('9', 27, '9');
        $emptyTimestamp = str_pad('9', 9, '9');

        while (strlen($emptySignatureFragment) < 2187) {
            $emptySignatureFragment = $emptySignatureFragment . '9';
        };

        for ($i = 0; $i < count($this->bundle); $i++) {

            // Fill empty signatureMessageFragment
            $this->bundle[$i]->signatureMessageFragment = $signatureFragments[$i] ? $signatureFragments[$i] : $emptySignatureFragment;

            // Fill empty trunkTransaction
            $this->bundle[$i]->trunkTransaction = $emptyHash;

            // Fill empty branchTransaction
            $this->bundle[$i]->branchTransaction = $emptyHash;

            $this->bundle[$i]->attachmentTimestamp = $emptyTimestamp;
            $this->bundle[$i]->attachmentTimestampLowerBound = $emptyTimestamp;
            $this->bundle[$i]->attachmentTimestampUpperBound = $emptyTimestamp;

            // Fill empty nonce
            $this->bundle[$i]->nonce = $emptyTag;
        }
    }


    public function finalize()
    {
        $validBundle = false;

        while (!$validBundle) {

            $kerl = new Kerl();
            $kerl->initialize();

            for ($i = 0; $i < count($this->bundle); $i++) {

                $valueTrits = $this->converter->trytes_to_trits($this->bundle[$i]->value);

                while (count($valueTrits) < 81) {
                    array_push($valueTrits, 0);
                }

                $timestampTrits = Converter . trits($this->bundle[$i]->timestamp);
                while (count($timestampTrits) < 27) {
                    array_push($timestampTrits, 0);
                }

                $currentIndexTrits = $this->converter->trytes_to_trits($this->bundle[$i]->currentIndex = $i);
                while (count($currentIndexTrits) < 27) {
                    array_push($currentIndexTrits, 0);
                }

                $lastIndexTrits = $this->converter->trytes_to_trits($this->bundle[$i]->lastIndex = count($this->bundle) - 1);
                while (count($lastIndexTrits) < 27) {
                    array_push($lastIndexTrits, 0);
                }

                $bundleEssence = $this->converter->trytes_to_trits(
                    $this->bundle[$i]->address .
                    $this->converter->trits_to_tryes($valueTrits) .
                    $this->bundle[$i]->obsoleteTag .
                    $this->converter->trits_to_tryes($timestampTrits) .
                    $this->converter->trits_to_tryes($currentIndexTrits) .
                    $this->converter->trits_to_tryes($lastIndexTrits));

                $kerl->absorb($bundleEssence, 0, count($bundleEssence));
            }

            $hash = [];
            $kerl->squeeze($hash, 0, HASH_LENGTH);
            $hash = $this->converter->trits_to_tryes($hash);

            for ($i = 0; $i < count($this->bundle); $i++) {

                $this->bundle[$i]->bundle = $hash;
            }

            $normalizedHash = $this->normalizedBundle($hash);
            if (in_array(13, $normalizedHash)) {
                // Insecure bundle. Increment Tag and recompute bundle hash.

                $tritArray = $this->converter->trytes_to_trits($this->bundle[0]->obsoleteTag);
                $increasedTag = Helper::tritAdd($tritArray, [1]);

                $this->bundle[0]->obsoleteTag = $this->converter->trits_to_tryes($increasedTag);
            } else {
                $validBundle = true;
            }
        }
    }

    public function normalizedBundle($bundleHash)
    {

        $normalizedBundle = [];

        for ($i = 0; $i < 3; $i++) {

            $sum = 0;
            for ($j = 0; $j < 27; $j++) {

                $tryteToConvertToTrits = $bundleHash{$i * 27 + $j};
                $tritsToConvertToInt = $this->converter->trytes_to_trits($tryteToConvertToTrits);
                $intsForNormalizedBundle = $this->converter->trits_to_integers($tritsToConvertToInt);
                $normalizedBundle[$i * 27 + $j] = $intsForNormalizedBundle;

                $sum = $sum + $normalizedBundle[$i * 27 + $j];
            }

            if ($sum >= 0) {

                while ($sum-- > 0) {

                    for ($j = 0; $j < 27; $j++) {

                        if ($normalizedBundle[$i * 27 + $j] > -13) {

                            $normalizedBundle[$i * 27 + $j]--;
                            break;
                        }
                    }
                }
            } else {

                while ($sum++ < 0) {

                    for ($j = 0; $j < 27; $j++) {

                        if ($normalizedBundle[$i * 27 + $j] < 13) {

                            $normalizedBundle[$i * 27 + $j]++;
                            break;
                        }
                    }
                }
            }
        }

        return $normalizedBundle;
    }
}


