<?php

require_once("Converter.php");
require_once("Kerl.php");
require_once("Bundle.php");
require_once("Helper.php");
//var oldSigning = require("./oldSigning");
//var errors = require("../../errors/inputErrors");


class Signing
{

    const HASH_LENGTH = 243;

    public function __construct()
    {
    }

    public static function key(&$seed, $index, $length)
    {
        while ((count($seed) % self::HASH_LENGTH) !== 0) {
            array_push($seed, 0);
        }

        $indexTrits = Converter::integer_to_trits($index);
        $subseed = Helper::tritAdd(array_slice($seed, 0), $indexTrits);

        $kerl = new Kerl();

        $kerl->initialize();
        $kerl->absorb($subseed, 0, count($subseed));
        $kerl->squeeze($subseed, 0, count($subseed));

        $kerl->reset();
        $kerl->absorb($subseed, 0, count($subseed));

        $key = [];
        $offset = 0;
        $buffer = [];

        while ($length-- > 0) {

            for ($i = 0; $i < 27; $i++) {

                $kerl->squeeze($buffer, 0, count($subseed));
                for ($j = 0; $j < 243; $j++) {
                    $key[$offset++] = $buffer[$j];
                }
            }
        }
        return $key;
    }

    /**
     *
     *
     **/
    public static function digests($key)
    {

        $digests = [];
        $buffer = [];

        for ($i = 0; $i < floor(count($key) / 6561); $i++) {

            $keyFragment = array_slice($key, $i * 6561, ($i + 1) * 6561);

            for ($j = 0; $j < 27; $j++) {

                $buffer = array_slice($keyFragment, $j * 243, ($j + 1) * 243);

                for ($k = 0; $k < 26; $k++) {

                    $kKerl = new Kerl();
                    $kKerl->initialize();
                    $kKerl->absorb($buffer, 0, count($buffer));
                    $kKerl->squeeze($buffer, 0, self::HASH_LENGTH);
                }

                for ($k = 0; $k < 243; $k++) {

                    $keyFragment[$j * 243 + $k] = $buffer[$k];
                }
            }

            $kerl = new Kerl();

            $kerl->initialize();
            $kerl->absorb($keyFragment, 0, count($keyFragment));
            $kerl->squeeze($buffer, 0, self::HASH_LENGTH);

            for ($j = 0; $j < 243; $j++) {

                $digests[$i * 243 + $j] = $buffer[$j];
            }
        }
        return $digests;
    }

    /**
     *
     *
     **/
    public static function address($digests)
    {

        $addressTrits = [];

        $kerl = new Kerl();

        $kerl->initialize();
        $kerl->absorb($digests, 0, count($digests));
        $kerl->squeeze($addressTrits, 0, self::HASH_LENGTH);

        return $addressTrits;
    }

    /**
     *
     *
     **/
    public static function digest($normalizedBundleFragment, $signatureFragment)
    {
        $buffer = [];

        $kerl = new Kerl();
        $kerl->initialize();

        for ($i = 0; $i < 27; $i++) {
            $buffer = array_slice($signatureFragment, $i * 243, ($i + 1) * 243);

            for ($j = $normalizedBundleFragment[$i] + 13; $j-- > 0;) {

                $jKerl = new Kerl();

                $jKerl->initialize();
                $jKerl->absorb($buffer, 0, count($buffer));
                $jKerl->squeeze($buffer, 0, self::HASH_LENGTH);
            }

            $kerl->absorb($buffer, 0, count($buffer));
        }

        $kerl->squeeze($buffer, 0, self::HASH_LENGTH);
        return $buffer;
    }

    /**
     *
     *
     **/
    public static function signatureFragment($normalizedBundleFragment, $keyFragment)
    {
        $signatureFragment = array_slice($keyFragment, 0);
        $hash = [];

        $kerl = new Kerl();

        for ($i = 0; $i < 27; $i++) {

            $hash = array_slice($signatureFragment, $i * 243, ($i + 1) * 243);

            for ($j = 0; $j < 13 - $normalizedBundleFragment[$i]; $j++) {
                $kerl->initialize();
                $kerl->reset();
                $kerl->absorb($hash, 0, count($hash));
                $kerl->squeeze($hash, 0, self::HASH_LENGTH);
            }

            for ($j = 0; $j < 243; $j++) {

                $signatureFragment[$i * 243 + $j] = $hash[$j];
            }
        }

        return $signatureFragment;
    }

    /**
     *
     **/
    public static function validateSignatures($expectedAddress, $signatureFragments, $bundleHash)
    {
        if (!$bundleHash) {
            echo "EXCEPTION -- INVALID BUNDLE HASH";
            //update when proper error handling implemented
            //throw new Exception("Invalid bundle hash!);
        }

        $bundle = new Bundle();

        $normalizedBundleFragments = [];
        $normalizedBundleHash = $bundle->normalizedBundle($bundleHash);

        // Split hash into 3 fragments
        for ($i = 0; $i < 3; $i++) {
            $normalizedBundleFragments[$i] = array_slice($normalizedBundleHash, $i * 27, ($i + 1) * 27);
        }

        // Get digests
        $digests = [];

        for ($i = 0; $i < count($signatureFragments); $i++) {

            $digestBuffer = self::digest($normalizedBundleFragments[$i % 3], Converter::trytes_to_trits($signatureFragments[$i]));

            for ($j = 0; $j < 243; $j++) {

                $digests[$i * 243 + $j] = $digestBuffer[$j];
            }
        }

        $address = Converter::trits_to_trytes(self::address($digests));

        return ($expectedAddress === $address);
    }
}
