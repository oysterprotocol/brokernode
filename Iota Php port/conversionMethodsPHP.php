<?php


require_once('operators.php');
require_once('helpers.php');
require_once('Object.php');
require_once('Function.php');
require_once('Global.php');
require_once('Array.php');
require_once('Date.php');
require_once('RegExp.php');
require_once('String.php');
require_once('Number.php');
require_once('Boolean.php');
require_once('Exception.php');

$RADIX = 3.0;
$RADIX_BYTES = 256.0;
$MAX_TRIT_VALUE = 1.0;
$MIN_TRIT_VALUE = -1.0;
$BYTE_HASH_LENGTH = 48.0;
$trytesAlphabet = "9ABCDEFGHIJKLMNOPQRSTUVWXYZ";
$trytesTrits = new Arr(new Arr(0.0, 0.0, 0.0), new Arr(1.0, 0.0, 0.0), new Arr(-1.0, 1.0, 0.0), new Arr(0.0, 1.0, 0.0), new Arr(1.0, 1.0, 0.0), new Arr(-1.0, -1.0, 1.0), new Arr(0.0, -1.0, 1.0), new Arr(1.0, -1.0, 1.0), new Arr(-1.0, 0.0, 1.0), new Arr(0.0, 0.0, 1.0), new Arr(1.0, 0.0, 1.0), new Arr(-1.0, 1.0, 1.0), new Arr(0.0, 1.0, 1.0), new Arr(1.0, 1.0, 1.0), new Arr(-1.0, -1.0, -1.0), new Arr(0.0, -1.0, -1.0), new Arr(1.0, -1.0, -1.0), new Arr(-1.0, 0.0, -1.0), new Arr(0.0, 0.0, -1.0), new Arr(1.0, 0.0, -1.0), new Arr(-1.0, 1.0, -1.0), new Arr(0.0, 1.0, -1.0), new Arr(1.0, 1.0, -1.0), new Arr(-1.0, -1.0, 0.0), new Arr(0.0, -1.0, 0.0), new Arr(1.0, -1.0, 0.0), new Arr(-1.0, 0.0, 0.0));
$trits = new Func(function($input = null, $state = null) use (&$Number, &$Math, &$trytesAlphabet, &$trytesTrits) {
  $trits = (is($or_ = $state) ? $or_ : new Arr());
  if (is(call_method($Number, "isInteger", $input))) {
    $absoluteValue = $input < 0.0 ? _negate($input) : $input;
    while ($absoluteValue > 0.0) {
      $remainder = (float)(to_number($absoluteValue) % 3.0);
      $absoluteValue = call_method($Math, "floor", _divide($absoluteValue, 3.0));
      if ($remainder > 1.0) {
        $remainder = -1.0;
        $absoluteValue++;
      }
      set($trits, get($trits, "length"), $remainder);
    }
    if ($input < 0.0) {
      for ($i = 0.0; $i < get($trits, "length"); $i++) {
        set($trits, $i, _negate(get($trits, $i)));
      }
    }
  } else {
    for ($i = 0.0; $i < get($input, "length"); $i++) {
      $index = call_method($trytesAlphabet, "indexOf", call_method($input, "charAt", $i));
      set($trits, to_number($i) * 3.0, get(get($trytesTrits, $index), 0.0));
      set($trits, _plus(to_number($i) * 3.0, 1.0), get(get($trytesTrits, $index), 1.0));
      set($trits, _plus(to_number($i) * 3.0, 2.0), get(get($trytesTrits, $index), 2.0));
    }
  }

  return $trits;
});
$trytes = new Func(function($trits = null) use (&$trytesAlphabet, &$trytesTrits) {
  $trytes = "";
  for ($i = 0.0; $i < get($trits, "length"); $i = _plus($i, 3.0)) {
    for ($j = 0.0; $j < get($trytesAlphabet, "length"); $j++) {
      if (get(get($trytesTrits, $j), 0.0) === get($trits, $i) && get(get($trytesTrits, $j), 1.0) === get($trits, _plus($i, 1.0)) && get(get($trytesTrits, $j), 2.0) === get($trits, _plus($i, 2.0))) {
        $trytes = _plus($trytes, call_method($trytesAlphabet, "charAt", $j));
        break;
      }
    }
  }
  return $trytes;
});
$value = new Func(function($trits = null) {
  $returnValue = 0.0;
  for ($i = get($trits, "length"); $i-- > 0.0; ) {
    $returnValue = _plus(to_number($returnValue) * 3.0, get($trits, $i));
  }
  return $returnValue;
});
$fromValue = new Func(function($value = null) use (&$RADIX, &$Math, &$MAX_TRIT_VALUE, &$MIN_TRIT_VALUE) {
  $destination = new Arr();
  $absoluteValue = $value < 0.0 ? _negate($value) : $value;
  $i = 0.0;
  while ($absoluteValue > 0.0) {
    $remainder = (float)(to_number($absoluteValue) % to_number($RADIX));
    $absoluteValue = call_method($Math, "floor", _divide($absoluteValue, $RADIX));
    if ($remainder > $MAX_TRIT_VALUE) {
      $remainder = $MIN_TRIT_VALUE;
      $absoluteValue++;
    }
    set($destination, $i, $remainder);
    $i++;
  }
  if ($value < 0.0) {
    for ($j = 0.0; $j < get($destination, "length"); $j++) {
      set($destination, $j, get($destination, $j) === 0.0 ? 0.0 : _negate(get($destination, $j)));
    }
  }
  return $destination;
});
$input = "G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA";
$trits = call($trits, $input);
echo trits;
