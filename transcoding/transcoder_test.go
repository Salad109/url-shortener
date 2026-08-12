package transcoding

import "testing"

func TestRoundTrip(t *testing.T) {
	ids := []int64{1, 2, 42, 1000, 123_456_789, MaxValue - 1}

	for _, id := range ids {
		encoded := Encode(id)
		if len(encoded) > MaxCodeLength {
			t.Errorf("Encode(%d) = %q, longer than %d characters", id, encoded, MaxCodeLength)
		}

		decoded, err := Decode(encoded)
		if err != nil {
			t.Errorf("Decode(%q) from id %d: %v", encoded, id, err)
			continue
		}
		if decoded != id {
			t.Errorf("Decode(Encode(%d)) = %d, want %d", id, decoded, id)
		}
	}
}

func TestEncodeIsDistinct(t *testing.T) {
	seen := make(map[string]int64)

	for id := int64(1); id <= 1000; id++ {
		encoded := Encode(id)
		if other, ok := seen[encoded]; ok {
			t.Fatalf("Encode(%d) = Encode(%d) = %q", id, other, encoded)
		}
		seen[encoded] = id
	}
}

func FuzzEncodeThenDecode(f *testing.F) {
	f.Add(int64(1))
	f.Add(int64(123_456_789))
	f.Add(int64(MaxValue - 1))

	f.Fuzz(func(t *testing.T, id int64) {
		if id < 1 || id >= MaxValue {
			t.Skip()
		}

		encoded := Encode(id)
		if len(encoded) > MaxCodeLength {
			t.Fatalf("Encode(%d) = %q, longer than %d characters", id, encoded, MaxCodeLength)
		}

		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Decode(%q) from id %d: %v", encoded, id, err)
		}
		if decoded != id {
			t.Errorf("Decode(Encode(%d)) = %d, want %d", id, decoded, id)
		}
	})
}

func FuzzDecodeThenEncode(f *testing.F) {
	f.Add("AB")
	f.Add("kVOkZ")

	f.Fuzz(func(t *testing.T, code string) {
		id, err := Decode(code)
		if err != nil {
			t.Skip()
		}

		if got := Encode(id); got != code {
			t.Errorf("Decode(%q) = %d, which re-encodes to %q", code, id, got)
		}
	})
}

func TestDecodeRejectsBadCodes(t *testing.T) {
	cases := map[string]string{
		"too long":     "abcdef",
		"invalid char": "ab!de",
		"out of range": "zzzzz",
		"non-ascii":    "abcó",
		"empty":        "",
		"leading zero": "0AB",
		"all zeroes":   "00000",
		"chinese":      "豆子",
		"emoji":        "🥺",
		"space":        "AB CD",
	}

	for name, code := range cases {
		if _, err := Decode(code); err == nil {
			t.Errorf("Decode(%q) (%s) returned no error", code, name)
		}
	}
}
