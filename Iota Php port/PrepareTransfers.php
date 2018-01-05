<?php

require_once('Converter.php');
require_once('InputValidator.php');
require_once('Utils.php');
require_once('Bundle.php');

class PrepareTransfers
{

    public function __construct()
    {

    }

    public static function prepareTransfers($seed, $transfers, $options = null, $callback)
    {
        $addHMAC = false;
        $addedHMAC = false;

        // validate the seed
        if (!InputValidator::isTrytes($seed)) {

            return call_user_func($callback, "Invalid Seed provided", null);
        }

        // If message or tag is not supplied, provide it
        // Also remove the checksum of the address if it's there after validating it
        foreach ($transfers as $thisTransfer) {

            $thisTransfer->message = $thisTransfer->message ? $thisTransfer->message : '';
            $thisTransfer->obsoleteTag = $thisTransfer->tag ? $thisTransfer->tag : ($thisTransfer->obsoleteTag ? $thisTransfer->obsoleteTag : '');

            // If address with checksum, validate it
            if (strlen($thisTransfer->address) === 90) {

                if (!Utils::isValidChecksum($thisTransfer->address)) {

                    return call_user_func($callback,
                        "Invalid Checksum supplied for address: " . $thisTransfer->address, null);

                }
            }

            $thisTransfer->address = Utils::noChecksum($thisTransfer->address);
        }

        // Input validation of transfers object
        if (!InputValidator::isTransfersArray($transfers)) {
            return call_user_func($callback, "Invalid transfers object", null);
        }

        $security = 2;

        // Create a new bundle
        $bundle = new Bundle();

        $totalValue = 0;
        $signatureFragments = [];
        $tag;

        //
        //  Iterate over all transfers, get totalValue
        //  and prepare the signatureFragments, message and tag
        //
        for ($i = 0; $i < count($transfers); $i++) {

            $signatureMessageLength = 1;

            // If message longer than 2187 trytes, increase signatureMessageLength (add 2nd transaction)
            if (count($transfers[$i]->message) > 2187) {

                // Get total length, message / maxLength (2187 trytes)
                $signatureMessageLength += intval(floor(count($transfers[$i]->message) / 2187));

                $msgCopy = $transfers[$i]->message;

                // While there is still a message, copy it
                while ($msgCopy) {

                    $fragment = substr($msgCopy, 0, 2187);
                    $msgCopy = substr($msgCopy, 2187, count($msgCopy));

                    // Pad remainder of fragment
                    while (is_null($fragment) || count($fragment) < 2187) {
                        $fragment .= '9';
                    }

                    $signatureFragments[] = $fragment;
                }
            } else {
                // Else, get single fragment with 2187 of 9's trytes
                $fragment = '';

                if ($transfers[$i]->message) {
                    $fragment = substr($transfers[$i]->message, 0, 2187);
                }

                while (strlen($fragment) < 2187) {
                    $fragment .= '9';
                };

                $signatureFragments[] = $fragment;
            }

            // get current timestamp in seconds
            $timestamp = intval(floor(time()));
            //$timestamp = 1511111111;
            //use this timestamp for comparing with output of other libraries

            // If no tag defined, get 27 tryte tag.
            $tag = $transfers[$i]->obsoleteTag ? $transfers[$i]->obsoleteTag : '999999999999999999999999999';

            // Pad for required 27 tryte length
            while (strlen($tag) < 27) {
                $tag .= '9';
            };

            // Add first entries to the bundle
            // Slice the address in case the user provided a checksummed one
            $bundle->addEntry($signatureMessageLength, $transfers[$i]->address, $transfers[$i]->value, $tag, $timestamp);
            // Sum up total value
            $totalValue += intval($transfers[$i]->value);
        }

        // If no input required, don't sign and simply finalize the bundle
        $bundle->finalize();
        $bundle->addTrytes($signatureFragments);

        $bundleTrytes = [];

        foreach ($bundle->bundle as $tx) {
            $bundleTrytes[] = Utils::transactionTrytes($tx);
        };

        return call_user_func($callback, null, array_reverse($bundleTrytes));
    }
}

$seed = 'A9YZ9YMXBQRBKKQYLZSDPIBPWLOURJPQHQDSOE9QBAC9XYABCIMNWPWMX9NVCDSWOTMIWSMDJRFWPDSKC';
$addressWithoutChecksum = 'SSEWOZSDXOVIURQRBTBDLQXWIXOLEUXHYBGAVASVPZ9HBTYJJEWBR9PDTGMXZGKPTGSUDW9QLFPJHTIEQ';
$addressWithChecksum = 'SSEWOZSDXOVIURQRBTBDLQXWIXOLEUXHYBGAVASVPZ9HBTYJJEWBR9PDTGMXZGKPTGSUDW9QLFPJHTIEQZNXDGNRJE';

$transactionObject = new stdClass();

$transactionObject->address = $addressWithoutChecksum;
$transactionObject->value = 0;
$transactionObject->message = 'WBTCGDGDPCVCTC';
$transactionObject->tag = 'CCPCVC';

PrepareTransfers::prepareTransfers($seed,
    [$transactionObject],
    null, //where options with inputs array would go, consider removing this and removing param from method
    function ($e, $s) {
        if ($s != null) {
            echo implode(", ", $s);
            //do something with $s, which should be an array of transaction trytes
        } else {
            echo "did not work";
            //do something with this error
        }
    });