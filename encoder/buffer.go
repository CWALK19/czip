package encoder

import "bytes"

type Indexes struct {
	AddressIndexes map[string]uint
	Bytes32Indexes map[string]uint
	Bytes4Indexes  map[string]uint
}

type Buffer struct {
	Commited []byte
	Pending  []byte

	Refs *References
}

type References struct {
	useContractStorage bool

	Indexes *Indexes

	usedFlags        map[string]int
	usedStorageFlags map[string]int
}

func NewBuffer(indexes *Indexes, useStorage bool) *Buffer {
	return &Buffer{
		// Start with an empty byte, this
		// will be used as the method when calling the compressor
		// contract.
		Commited: make([]byte, 1),
		Pending:  make([]byte, 0),

		Refs: &References{
			Indexes:            indexes,
			useContractStorage: useStorage,
			usedFlags:          make(map[string]int),
			usedStorageFlags:   make(map[string]int),
		},
	}
}

func (r *References) Copy() *References {
	usedFlags := make(map[string]int, len(r.usedFlags))
	for k, v := range r.usedFlags {
		usedFlags[k] = v
	}

	usedStorageFlags := make(map[string]int, len(r.usedStorageFlags))
	for k, v := range r.usedStorageFlags {
		usedStorageFlags[k] = v
	}

	return &References{
		useContractStorage: r.useContractStorage,

		usedFlags:        usedFlags,
		usedStorageFlags: usedStorageFlags,
	}
}

func (cb *Buffer) Data() []byte {
	return cb.Commited
}

func (cb *Buffer) Len() int {
	return len(cb.Commited)
}

func (cb *Buffer) WriteByte(b byte) {
	cb.Pending = append(cb.Pending, b)
}

func (cb *Buffer) WriteBytes(b []byte) {
	cb.Pending = append(cb.Pending, b...)
}

func (cb *Buffer) WriteInt(i uint) {
	cb.WriteByte(byte(i))
}

func (cb *Buffer) FindPastData(data []byte) int {
	for i := 0; i+len(data) < len(cb.Commited); i++ {
		if bytes.Equal(cb.Commited[i:i+len(data)], data) {
			return i
		}
	}

	return -1
}

func (cb *Buffer) End(uncompressed []byte, t EncodeType) {
	// We need 2 bytes to point to a flag, so any uncompressed value
	// that is 2 bytes or less is not worth saving.
	if len(uncompressed) > 2 {
		rindex := cb.Len()

		switch t {
		case ReadStorage:
		case Stateless:
			cb.Refs.usedFlags[string(uncompressed)] = rindex + 1
		case WriteStorage:
			cb.Refs.usedStorageFlags[string(uncompressed)] = rindex + 1
		default:
		}
	}

	cb.Commited = append(cb.Commited, cb.Pending...)
	cb.Pending = nil
}

type Snapshot struct {
	Commited []byte

	SignatureLevel uint

	Refs *References
}

func (cb *Buffer) Snapshot() *Snapshot {
	// Create a copy of the commited buffer
	// and of the references.
	com := make([]byte, len(cb.Commited))
	copy(com, cb.Commited)

	refs := cb.Refs.Copy()

	return &Snapshot{
		Commited: com,
		Refs:     refs,
	}
}

func (cb *Buffer) Restore(snap *Snapshot) {
	cb.Commited = snap.Commited
	cb.Refs = snap.Refs
}
