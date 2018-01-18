<?php

class Helper
{

    public function __construct()
    {
    }


    protected static function sum($a, $b)
    {
        $s = $a + $b;

        switch ($s) {

            case 2:
                return -1;
            case -2:
                return 1;
            default:
                return $s;
        }
    }

    protected static function cons($a, $b)
    {

        if ($a === $b) {

            return $a;

        }

        return 0;
    }

    protected static function any($a, $b)
    {

        $s = $a + $b;

        if ($s > 0) {

            return 1;

        }
        if ($s < 0) {

            return -1;

        }
        return 0;
    }

    protected static function full_add($a, $b, $c)
    {

        $s_a = self::sum($a, $b);
        $c_a = self::cons($a, $b);
        $c_b = self::cons($s_a, $c);
        $c_out = self::any($c_a, $c_b);
        $s_out = self::sum($s_a, $c);

        $retVal = [$s_out, $c_out];
        return $retVal;

    }

    public static function tritAdd($a, $b)
    {

        $length = max(count($a), count($b));
        $out = [];
        $carry = 0;

        for ($i = 0; $i < $length; $i++) {

            $a_i = $i < count($a) ? $a[$i] : 0;
            $b_i = $i < count($b) ? $b[$i] : 0;
            $f_a = self::full_add($a_i, $b_i, $carry);
            $out[$i] = $f_a[0];
            $carry = $f_a[1];

        }
        return $out;
    }
}

