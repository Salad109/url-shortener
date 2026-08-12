package transcoding

import "testing"

func TestRoundTrip(t *testing.T) {
	ids := []int64{1, 2, 42, 1000, 123_456_789, MaxID}

	for _, id := range ids {
		encoded, err := Encode(id)
		if err != nil {
			t.Errorf("Encode(%d): %v", id, err)
			continue
		}
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
		encoded, err := Encode(id)
		if err != nil {
			t.Fatalf("Encode(%d): %v", id, err)
		}
		if other, ok := seen[encoded]; ok {
			t.Fatalf("Encode(%d) = Encode(%d) = %q", id, other, encoded)
		}
		seen[encoded] = id
	}
}

func FuzzEncodeThenDecode(f *testing.F) {
	f.Add(int64(1))
	f.Add(int64(123_456_789))
	f.Add(int64(MaxID))

	f.Fuzz(func(t *testing.T, id int64) {
		if id < 1 || id > MaxID {
			t.Skip()
		}

		encoded, err := Encode(id)
		if err != nil {
			t.Fatalf("Encode(%d): %v", id, err)
		}
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

		got, err := Encode(id)
		if err != nil {
			t.Fatalf("Decode(%q) = %d, which Encode rejects: %v", code, id, err)
		}
		if got != code {
			t.Errorf("Decode(%q) = %d, which re-encodes to %q", code, id, got)
		}
	})
}

func TestIDCeiling(t *testing.T) {
	encoded, err := Encode(MaxID)
	if err != nil {
		t.Fatalf("Encode(MaxID): %v", err)
	}
	if decoded, err := Decode(encoded); err != nil || decoded != MaxID {
		t.Errorf("Encode(MaxID) round trips to %d, %v, want %d and no error", decoded, err, MaxID)
	}

	for _, id := range []int64{0, -1, MaxID + 1, MaxID + 2} {
		if _, err := Encode(id); err == nil {
			t.Errorf("Encode(%d) returned no error", id)
		}
	}
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
