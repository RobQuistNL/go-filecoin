package consensus

// #cgo LDFLAGS: -L${SRCDIR}/../proofs/lib -lfilecoin_proofs
// #cgo pkg-config: ${SRCDIR}/../proofs/lib/pkgconfig/libfilecoin_proofs.pc
// #include "../proofs/include/libfilecoin_proofs.h"
import "C"

func flarp() uint64 {
	return 128823
}
