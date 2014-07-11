package tools

import (
	"encoding/hex"
	"log"
	"testing"

	"github.com/conformal/btcec"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSign(t *testing.T) {
	msg := []byte("test message")
	hexPrivkey := "22a47fa09a223f2aa079edf85a7c2d4f8720ee63e502ee2869afab7de234b80c"

	// Decode the hex-encoded private key.
	bytesPrivkey, err := hex.DecodeString(hexPrivkey)
	if err != nil {
		log.Fatal(err)
	}

	priv, pub := btcec.PrivKeyFromBytes(btcec.S256(), bytesPrivkey)

	Convey("btcec example test", t, func() {
		sig, err := Sign(msg, priv)
		if err != nil {
			log.Fatal(err)
		}

		// Verify the signature for the message using the public key.
		So(Verify(msg, sig, pub), ShouldBeTrue)
	})
}

func TestVerify(t *testing.T) {
	msg := []byte("test message")
	hexPub := "02a673638cb9587cb68ea08dbef685c6f2d2a751a8b3c6f2a7e9a4999e6e4bfaf5"
	hexSig := "30450220090ebfb3690a0ff115bb1b38b8b323a667b7653454f1bccb06d4bbdca42c2079022100ec95778b51e7071cb1205f8bde9af6592fc978b0452dafe599481c46d6b2e479"

	// Decode hex-encoded serialized public key.
	bytesPub, err := hex.DecodeString(hexPub)
	if err != nil {
		log.Fatal(err)
	}
	pub, err := btcec.ParsePubKey(bytesPub, btcec.S256())
	if err != nil {
		log.Fatal(err)
	}

	// Decode the hex-encoded signature.
	bytesSig, err := hex.DecodeString(hexSig)
	if err != nil {
		log.Fatal(err)
	}
	sig, err := btcec.ParseSignature(bytesSig, btcec.S256())
	if err != nil {
		log.Fatal(err)
	}

	Convey("btcec example test", t, func() {
		// Verify the signature for the message using the public key.
		So(Verify(msg, sig, pub), ShouldBeTrue)
	})
}

func TestParseKeyPair(t *testing.T) {
	h := DetHashString("1")
	_, pub := ParseKeyPair(h)
	_, err := btcec.ParsePubKey(pub.SerializeUncompressed(), btcec.S256())

	Convey("PrivateKey='1'", t, func() {
		So(err, ShouldBeNil)
	})
}

type ZerosLeftTest struct {
	In   string
	Size int
	Out  string
}

func TestZerosLeft(t *testing.T) {
	desc := []string{
		"Empty string + size=-1",
		"Empty string + size=0",
		"Empty string + size=8",
		"len(s)=7 + size=-1",
		"len(s)=7 + size=0",
		"len(s)=7 + size=1",
		"len(s)=7 + size=2",
		"len(s)=7 + size=3",
		"len(s)=7 + size=4",
		"len(s)=7 + size=5",
		"len(s)=7 + size=6",
		"len(s)=7 + size=7",
		"len(s)=7 + size=8",
		"len(s)=7 + size=9",
		"len(unicode_s)=3 + size=-1",
		"len(unicode_s)=3 + size=0",
		"len(unicode_s)=3 + size=1",
		"len(unicode_s)=3 + size=2",
		"len(unicode_s)=3 + size=3",
		"len(unicode_s)=3 + size=4",
		"len(unicode_s)=3 + size=5",
	}

	cases := []ZerosLeftTest{
		ZerosLeftTest{"", -1, ""},
		ZerosLeftTest{"", 0, ""},
		ZerosLeftTest{"", 8, "00000000"},
		ZerosLeftTest{"altcoin", -1, "altcoin"},
		ZerosLeftTest{"altcoin", 0, "altcoin"},
		ZerosLeftTest{"altcoin", 1, "altcoin"},
		ZerosLeftTest{"altcoin", 2, "altcoin"},
		ZerosLeftTest{"altcoin", 3, "altcoin"},
		ZerosLeftTest{"altcoin", 4, "altcoin"},
		ZerosLeftTest{"altcoin", 5, "altcoin"},
		ZerosLeftTest{"altcoin", 6, "altcoin"},
		ZerosLeftTest{"altcoin", 7, "altcoin"},
		ZerosLeftTest{"altcoin", 8, "0altcoin"},
		ZerosLeftTest{"altcoin", 9, "00altcoin"},
		ZerosLeftTest{"世", -1, "世"},
		ZerosLeftTest{"世", 0, "世"},
		ZerosLeftTest{"世", 1, "世"},
		ZerosLeftTest{"世", 2, "世"},
		ZerosLeftTest{"世", 3, "世"},
		ZerosLeftTest{"世", 4, "0世"},
		ZerosLeftTest{"世", 5, "00世"},
	}

	for index, test := range cases {
		Convey(desc[index], t, func() {
			So(ZerosLeft(test.In, test.Size), ShouldEqual, test.Out)
		})
	}
}

type InTest struct {
	What       string
	Collection []string
	Out        bool
}

func TestIn(t *testing.T) {
	desc := []string{
		"Empty what and collection",
		"Empty collection",
		"Empty string",
		"Check true",
		"Check false",
		"Check true chinese (unicode)",
		"Check false chinese (unicode)",
	}

	cases := []InTest{
		InTest{"", []string{}, false},
		InTest{"altcoin", []string{}, false},
		InTest{"", []string{""}, true},
		InTest{"altcoin", []string{"bitcoin", "litecoin", "altcoin"}, true},
		InTest{"altcoin", []string{"bitcoin", "litecoin"}, false},
		InTest{"世界", []string{"Hello", "世界"}, true},
		InTest{"世", []string{"Hello", "世界"}, false},
	}

	for index, test := range cases {
		Convey(desc[index], t, func() {
			So(In(test.What, test.Collection), ShouldEqual, test.Out)
		})
	}
}

func TestNotIn(t *testing.T) {
	desc := []string{
		"Empty what and collection",
		"Empty collection",
		"Empty string",
		"Check true",
		"Check false",
		"Check true chinese (unicode)",
		"Check false chinese (unicode)",
	}

	cases := []InTest{
		InTest{"", []string{}, false},
		InTest{"altcoin", []string{}, false},
		InTest{"", []string{""}, true},
		InTest{"altcoin", []string{"bitcoin", "litecoin", "altcoin"}, true},
		InTest{"altcoin", []string{"bitcoin", "litecoin"}, false},
		InTest{"世界", []string{"Hello", "世界"}, true},
		InTest{"世", []string{"Hello", "世界"}, false},
	}

	for index, test := range cases {
		Convey(desc[index], t, func() {
			So(NotIn(test.What, test.Collection), ShouldEqual, !test.Out)
		})
	}
}

type JSONLenTest struct {
	In  interface{}
	Out int
}

func TestJSONLen(t *testing.T) {
	desc := []string{
		"Empty string",
		"Project's name",
		"world in chinese (unicode)",
		"Float number",
		"Error: channels",
		"Error: func",
		"Error: complex",
	}

	cases := []JSONLenTest{
		JSONLenTest{"", 2},
		JSONLenTest{"altcoin", 9},
		JSONLenTest{"世界", 8},
		JSONLenTest{3.16, 4},
		JSONLenTest{make(chan int), -1},
		JSONLenTest{func() {}, -1},
		JSONLenTest{2 + 3i, -1},
	}

	for index, test := range cases {
		Convey(desc[index], t, func() {
			So(JSONLen(test.In), ShouldEqual, test.Out)
		})
	}
}

type MaxTest struct {
	InA, InB int
	Out      int
}

func TestMax(t *testing.T) {
	desc := []string{
		"Max(-1, -1) = -1",
		"Max(-1, 0) = 0",
		"Max(0, -1) = 0",
		"Max(-2, -1) = -1",
		"Max(-1, -2) = -1",
		"Max(2, 1) = 2",
		"Max(1, 2) = 2",
		"Max(1 << 20, 1 << 21) = 1 << 21",
		"Max(1 << 21, 1 << 20) = 1 << 21",
	}

	cases := []MaxTest{
		MaxTest{-1, -1, -1},
		MaxTest{-1, 0, 0},
		MaxTest{0, -1, 0},
		MaxTest{-2, -1, -1},
		MaxTest{-1, -2, -1},
		MaxTest{2, 1, 2},
		MaxTest{1, 2, 2},
		MaxTest{1 << 20, 1 << 21, 1 << 21},
		MaxTest{1 << 21, 1 << 20, 1 << 21},
	}

	for index, test := range cases {
		Convey(desc[index], t, func() {
			So(Max(test.InA, test.InB), ShouldEqual, test.Out)
		})
	}
}
