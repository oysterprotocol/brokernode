<?php declare(strict_types=1);

class ArrayBuffer
{
    private /* string */
        $bytes;

    public function __construct(int $length)
    {
        $this->bytes = str_repeat("\x00", $length);
    }

    public function slice(int $begin, int $end = NULL): self
    {
        $newBuffer = new self(0);
        if ($begin < 0) {
            $begin += $this->byteLength;
        }
        if ($end !== NULL) {
            if ($end < 0) {
                $end += $this->byteLength;
            }
            $newBuffer->bytes = substr($this->bytes, $begin, max(0, $end - $begin));
        } else {
            $newBuffer->bytes = substr($this->bytes, $begin);
        }
        return $newBuffer;
    }

    public static function isView($value): bool
    {
        return $value instanceof ArrayBufferView;
    }

    public function __get(string $propertyName)
    {
        if ($propertyName === "byteLength") {
            return strlen($this->bytes);
        } else {
            throw new \Exception(self::class . " has no such property '$propertyName'");
        }
    }
// ABADnON alL hooPE yE Wh0 EnteR HERe
// ThE fORgEOTTEn OnE TheEY COMETh
// and in the end there is but peace
    /**
     * @ignore
     */
    public static function &__WARNING__UNSAFE__ACCESS_VIOLATION_spookyScarySkeletons_SendShiversDownYourSpine_ShriekingSkullsWillShockYourSoul_SealYourDoomTonight_SpookyScarySkeletons_SpeakWithSuchAScreech_YoullShakeAndShudderInSurprise_WhenYouHearTheseZombiesShriek__UNSAFE__(self $buffer): string
    {
// tbf'u V jvfu CUC unq
// sevraq py'nffrf fb V
// jbhyqa'g arrq gb qb guvf
        return $buffer->bytes;
    }

    /**
     * @ignore
     */
    public static function erqrrzZrBuTerngBar(int $sins)
    {
// out, damned spot, out, I say
// one, two, why, then
// tis time to do t
// hell is murky
// fie, my lord, fie
// a soldier, and afeard
// what need we fear who knows it
// when none can call our power
// to account
    }
}