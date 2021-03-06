package eval_test

import (
	"testing"

	. "github.com/elves/elvish/pkg/eval/evaltest"
)

func TestStringComparisonCommands(t *testing.T) {
	Test(t,
		That(`<s a b`).Puts(true),
		That(`<s 2 10`).Puts(false),
		That(`<=s a a`).Puts(true),
		That(`<=s a b`).Puts(true),
		That(`<=s b a`).Puts(false),
		That(`==s haha haha`).Puts(true),
		That(`==s 10 10.0`).Puts(false),
		That(`!=s haha haha`).Puts(false),
		That(`!=s 10 10.1`).Puts(true),
		That(`>s a b`).Puts(false),
		That(`>s 2 10`).Puts(true),
		That(`>=s a a`).Puts(true),
		That(`>=s a b`).Puts(false),
		That(`>=s b a`).Puts(true),
	)
}

func TestToString(t *testing.T) {
	Test(t,
		That(`to-string str (float64 1) $true`).Puts("str", "1", "$true"),
	)
}

func TestOrd(t *testing.T) {
	Test(t,
		That(`ord a`).Puts("0x61"),
		That(`ord 你好`).Puts("0x4f60", "0x597d"),
	)
}

func TestChr(t *testing.T) {
	Test(t,
		That(`chr 0x61`).Puts("a"),
		That(`chr 0x4f60 0x597d`).Puts("你好"),
		That(`chr -1`).Throws(AnyError),
	)
}

func TestBase(t *testing.T) {
	Test(t,
		That(`base 2 1 3 4 16 255`).Puts("1", "11", "100", "10000", "11111111"),
		That(`base 16 42 233`).Puts("2a", "e9"),
		That(`base 1 1`).Throws(AnyError),   // no base-1
		That(`base 37 10`).Throws(AnyError), // no letter for base-37
	)
}

func TestWcswidth(t *testing.T) {
	Test(t,
		That(`wcswidth 你好`).Puts("4"),
		That(`-override-wcwidth x 10; wcswidth 1x2x; -override-wcwidth x 1`).
			Puts("22"),
	)
}

func TestHasPrefix(t *testing.T) {
	Test(t,
		That(`has-prefix golang go`).Puts(true),
		That(`has-prefix golang x`).Puts(false),
	)
}

func TestHasSuffix(t *testing.T) {
	Test(t,
		That(`has-suffix golang x`).Puts(false),
	)
}

func TestEawk(t *testing.T) {
	Test(t,
		That(`echo "  ax  by cz  \n11\t22 33" | eawk [@a]{ put $a[-1] }`).
			Puts("cz", "33"),
	)
}
