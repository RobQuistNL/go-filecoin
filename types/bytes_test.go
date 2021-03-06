package types

import (
	"bytes"
	cbor "github.com/ipfs/go-ipld-cbor"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundtrip(t *testing.T) {
	assert := assert.New(t)
	cases := [][]byte{nil, {}, []byte("bytes")}
	for _, c := range cases {
		b, err := cbor.WrapObject(c, DefaultHashFunction, -1)
		assert.NoError(err)
		var out []byte
		err = cbor.DecodeInto(b.RawData(), &out)
		assert.NoError(err)
		switch {
		case c == nil:
			assert.Nil(out)
		default:
			assert.True(bytes.Equal(c, out))
		}
	}
}
