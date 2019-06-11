// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"errors"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

type DirectoryBlock struct {
	//Not Marshalized
	DBHash     interfaces.IHash `json:"dbhash"`
	KeyMR      interfaces.IHash `json:"keymr"`
	HeaderHash interfaces.IHash `json:"headerhash"`
	keyMRset   bool             `json:"keymrset"`

	//Marshalized
	Header    interfaces.IDirectoryBlockHeader `json:"header"`
	DBEntries []interfaces.IDBEntry            `json:"dbentries"`
}

var _ interfaces.Printable = (*DirectoryBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*DirectoryBlock)(nil)
var _ interfaces.IDirectoryBlock = (*DirectoryBlock)(nil)
var _ interfaces.DatabaseBatchable = (*DirectoryBlock)(nil)
var _ interfaces.DatabaseBlockWithEntries = (*DirectoryBlock)(nil)

func (d *DirectoryBlock) Init() {
	if d.Header == nil {
		h := new(DBlockHeader)
		h.Init()
		d.Header = h
	}
}

func (d *DirectoryBlock) IsSameAs(b interfaces.IDirectoryBlock) bool {
	if d == nil || b == nil {
		if d == nil && b == nil {
			return true
		}
		return false
	}

	if d.Header.IsSameAs(b.GetHeader()) == false {
		return false
	}
	bDBEntries := b.GetDBEntries()
	if len(d.DBEntries) != len(bDBEntries) {
		return false
	}
	for i := range d.DBEntries {
		if d.DBEntries[i].IsSameAs(bDBEntries[i]) == false {
			return false
		}
	}
	return true
}

func (d *DirectoryBlock) SetEntryHash(hash, chainID interfaces.IHash, index int) {
	if len(d.DBEntries) <= index {
		ent := make([]interfaces.IDBEntry, index+1)
		copy(ent, d.DBEntries)
		d.DBEntries = ent
	}
	dbe := new(DBEntry)
	dbe.ChainID = chainID
	dbe.KeyMR = hash
	d.DBEntries[index] = dbe
}

func (d *DirectoryBlock) SetABlockHash(aBlock interfaces.IAdminBlock) error {
	hash := aBlock.DatabasePrimaryIndex()
	d.SetEntryHash(hash, aBlock.GetChainID(), 0)
	return nil
}

func (d *DirectoryBlock) SetECBlockHash(ecBlock interfaces.IEntryCreditBlock) error {
	hash := ecBlock.DatabasePrimaryIndex()
	d.SetEntryHash(hash, ecBlock.GetChainID(), 1)
	return nil
}

func (d *DirectoryBlock) SetFBlockHash(fBlock interfaces.IFBlock) error {
	hash := fBlock.DatabasePrimaryIndex()
	d.SetEntryHash(hash, fBlock.GetChainID(), 2)
	return nil
}

func (d *DirectoryBlock) GetEntryHashes() []interfaces.IHash {
	entries := d.DBEntries[:]
	answer := make([]interfaces.IHash, len(entries))
	for i, entry := range entries {
		answer[i] = entry.GetKeyMR()
	}
	return answer
}

func (d *DirectoryBlock) GetEntrySigHashes() []interfaces.IHash {
	return nil
}

//bubble sort
func (d *DirectoryBlock) Sort() {
	done := false
	for i := 3; !done && i < len(d.DBEntries)-1; i++ {
		done = true
		for j := 3; j < len(d.DBEntries)-1-i+3; j++ {
			comp := bytes.Compare(d.DBEntries[j].GetChainID().Bytes(),
				d.DBEntries[j+1].GetChainID().Bytes())
			if comp > 0 {
				h := d.DBEntries[j]
				d.DBEntries[j] = d.DBEntries[j+1]
				d.DBEntries[j+1] = h
				done = false
			}
		}
	}
}

func (d *DirectoryBlock) GetEntryHashesForBranch() []interfaces.IHash {
	entries := d.DBEntries[:]
	answer := make([]interfaces.IHash, 2*len(entries))
	for i, entry := range entries {
		answer[2*i] = entry.GetChainID()
		answer[2*i+1] = entry.GetKeyMR()
	}
	return answer
}

func (d *DirectoryBlock) GetDBEntries() []interfaces.IDBEntry {
	return d.DBEntries
}

func (d *DirectoryBlock) GetEBlockDBEntries() []interfaces.IDBEntry {
	answer := []interfaces.IDBEntry{}
	for _, v := range d.DBEntries {
		if v.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000a" {
			continue
		}
		if v.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000f" {
			continue
		}
		if v.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000c" {
			continue
		}
		answer = append(answer, v)
	}
	return answer
}

func (d *DirectoryBlock) CheckDBEntries() error {
	if len(d.DBEntries) < 3 {
		return fmt.Errorf("Not enough entries - %v", len(d.DBEntries))
	}
	if d.DBEntries[0].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000a" {
		return fmt.Errorf("Invalid ChainID at position 0 - %v", d.DBEntries[0].GetChainID().String())
	}
	if d.DBEntries[1].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000c" {
		return fmt.Errorf("Invalid ChainID at position 1 - %v", d.DBEntries[1].GetChainID().String())
	}
	if d.DBEntries[2].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000f" {
		return fmt.Errorf("Invalid ChainID at position 2 - %v", d.DBEntries[2].GetChainID().String())
	}
	return nil
}

func (d *DirectoryBlock) GetKeyMR() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.GetKeyMR() saw an interface that was nil")
		}
	}()
	keyMR, err := d.BuildKeyMerkleRoot()
	if err != nil {
		panic("Failed to build the key MR")
	}

	//if d.keyMRset && d.KeyMR.Fixed() != keyMR.Fixed() {
	//	panic("keyMR changed!")
	//}

	d.KeyMR = keyMR
	d.keyMRset = true

	return d.KeyMR
}

func (d *DirectoryBlock) GetHeader() interfaces.IDirectoryBlockHeader {
	d.Init()
	return d.Header
}

func (d *DirectoryBlock) SetHeader(header interfaces.IDirectoryBlockHeader) {
	d.Header = header
}

func (d *DirectoryBlock) SetDBEntries(dbEntries []interfaces.IDBEntry) error {
	if dbEntries == nil {
		return errors.New("dbEntries cannot be nil")
	}

	d.DBEntries = dbEntries
	return nil
}

func (d *DirectoryBlock) New() interfaces.BinaryMarshallableAndCopyable {
	dBlock := new(DirectoryBlock)
	dBlock.Header = NewDBlockHeader()
	dBlock.DBHash = primitives.NewZeroHash()
	dBlock.KeyMR = primitives.NewZeroHash()
	return dBlock
}

func (d *DirectoryBlock) GetDatabaseHeight() uint32 {
	d.Init()
	return d.GetHeader().GetDBHeight()
}

func (d *DirectoryBlock) GetChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.GetChainID() saw an interface that was nil")
		}
	}()
	return primitives.NewHash(constants.D_CHAINID)
}

func (d *DirectoryBlock) DatabasePrimaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.DatabasePrimaryIndex() saw an interface that was nil")
		}
	}()
	return d.GetKeyMR()
}

func (d *DirectoryBlock) DatabaseSecondaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.DatabaseSecondaryIndex() saw an interface that was nil")
		}
	}()
	return d.GetHash()
}

func (d *DirectoryBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(d)
}

func (d *DirectoryBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(d)
}

func (d *DirectoryBlock) String() string {
	d.Init()
	var out primitives.Buffer

	kmr := d.GetKeyMR()
	out.WriteString(fmt.Sprintf("%20s %v\n", "keymr:", kmr.String()))

	kmr = d.BodyKeyMR()
	out.WriteString(fmt.Sprintf("%20s %v\n", "bodymr:", kmr.String()))

	fh := d.GetFullHash()
	out.WriteString(fmt.Sprintf("%20s %v\n", "fullhash:", fh.String()))

	out.WriteString(d.GetHeader().String())
	out.WriteString("entries:\n")
	for i, entry := range d.DBEntries {
		out.WriteString(fmt.Sprintf("%5d %s", i, entry.String()))
	}

	return (string)(out.DeepCopyBytes())
}

func (d *DirectoryBlock) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DirectoryBlock.MarshalBinary err:%v", *pe)
		}
	}(&err)
	d.Init()
	d.Sort()
	_, err = d.BuildBodyMR()
	if err != nil {
		return nil, err
	}

	buf := primitives.NewBuffer(nil)

	err = buf.PushBinaryMarshallable(d.GetHeader())
	if err != nil {
		return nil, err
	}

	for i := uint32(0); i < d.Header.GetBlockCount(); i++ {
		err = buf.PushBinaryMarshallable(d.GetDBEntries()[i])
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), err
}

func (d *DirectoryBlock) BuildBodyMR() (interfaces.IHash, error) {
	count := uint32(len(d.GetDBEntries()))
	d.GetHeader().SetBlockCount(count)
	if count == 0 {
		panic("Zero block size!")
	}

	hashes := make([]interfaces.IHash, len(d.GetDBEntries()))
	for i, entry := range d.GetDBEntries() {
		data, err := entry.MarshalBinary()
		if err != nil {
			return nil, err
		}
		hashes[i] = primitives.Sha(data)
	}

	if len(hashes) == 0 {
		hashes = append(hashes, primitives.Sha(nil))
	}

	merkleTree := primitives.BuildMerkleTreeStore(hashes)
	merkleRoot := merkleTree[len(merkleTree)-1]

	d.GetHeader().SetBodyMR(merkleRoot)

	return merkleRoot, nil
}

func (d *DirectoryBlock) GetHeaderHash() (interfaces.IHash, error) {
	d.Header.SetBlockCount(uint32(len(d.GetDBEntries())))
	return d.Header.GetHeaderHash()
}

func (d *DirectoryBlock) BodyKeyMR() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.BodyKeyMR() saw an interface that was nil")
		}
	}()
	key, _ := d.BuildBodyMR()
	return key
}

func (d *DirectoryBlock) BuildKeyMerkleRoot() (keyMR interfaces.IHash, err error) {
	// Create the Entry Block Key Merkle Root from the hash of Header and the Body Merkle Root

	hashes := make([]interfaces.IHash, 0, 2)
	bodyKeyMR := d.BodyKeyMR() //This needs to be called first to build the header properly!!
	headerHash, err := d.GetHeaderHash()
	if err != nil {
		return nil, err
	}
	hashes = append(hashes, headerHash)
	hashes = append(hashes, bodyKeyMR)
	merkle := primitives.BuildMerkleTreeStore(hashes)
	keyMR = merkle[len(merkle)-1] // MerkleRoot is not marshalized in Dir Block

	d.KeyMR = keyMR

	d.GetFullHash() // Create the Full Hash when we create the keyMR

	return primitives.NewHash(keyMR.Bytes()), nil
}

func UnmarshalDBlock(data []byte) (interfaces.IDirectoryBlock, error) {
	dBlock := new(DirectoryBlock)
	dBlock.Header = NewDBlockHeader()
	dBlock.DBHash = primitives.NewZeroHash()
	dBlock.KeyMR = primitives.NewZeroHash()
	err := dBlock.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return dBlock, nil
}

func (d *DirectoryBlock) UnmarshalBinaryData(data []byte) ([]byte, error) {
	newData := data
	var err error
	var fbh interfaces.IDirectoryBlockHeader = new(DBlockHeader)

	newData, err = fbh.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}
	d.SetHeader(fbh)

	// entryLimit is the maximum number of 32 byte entries that could fit in the body of the binary dblock
	entryLimit := uint32(len(newData) / 32)
	entryCount := d.GetHeader().GetBlockCount()
	if entryCount > entryLimit {
		return nil, fmt.Errorf(
			"Error: DirectoryBlock.UnmarshalBinary: Entry count %d is larger "+
				"than body size %d. (uint underflow?)",
			entryCount, entryLimit,
		)
	}

	entries := make([]interfaces.IDBEntry, entryCount)
	for i := uint32(0); i < entryCount; i++ {
		entries[i] = new(DBEntry)
		newData, err = entries[i].UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
	}

	err = d.SetDBEntries(entries)
	if err != nil {
		return nil, err
	}

	err = d.CheckDBEntries()
	if err != nil {
		return nil, err
	}

	return newData, nil
}

func (h *DirectoryBlock) GetTimestamp() interfaces.Timestamp {
	return h.GetHeader().GetTimestamp()
}

func (d *DirectoryBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = d.UnmarshalBinaryData(data)
	return
}

func (d *DirectoryBlock) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.GetHash() saw an interface that was nil")
		}
	}()
	return d.GetFullHash()
}

func (d *DirectoryBlock) GetFullHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirectoryBlock.GetFullHash() saw an interface that was nil")
		}
	}()
	binaryDblock, err := d.MarshalBinary()
	if err != nil {
		return nil
	}
	d.DBHash = primitives.Sha(binaryDblock)
	return d.DBHash
}

func (d *DirectoryBlock) AddEntry(chainID interfaces.IHash, keyMR interfaces.IHash) error {
	var dbentry interfaces.IDBEntry
	dbentry = new(DBEntry)
	dbentry.SetChainID(chainID)
	dbentry.SetKeyMR(keyMR)

	return d.SetDBEntries(append(d.DBEntries, dbentry))
}

/*********************************************************************
 * Support
 *********************************************************************/

func NewDirectoryBlock(prev interfaces.IDirectoryBlock) interfaces.IDirectoryBlock {
	newdb := new(DirectoryBlock)

	newdb.Header = new(DBlockHeader)
	newdb.GetHeader().SetVersion(constants.VERSION_0)

	if prev != nil {
		newdb.GetHeader().SetPrevFullHash(prev.GetFullHash())
		newdb.GetHeader().SetPrevKeyMR(prev.GetKeyMR())
		newdb.GetHeader().SetDBHeight(prev.GetHeader().GetDBHeight() + 1)
	} else {
		newdb.GetHeader().SetPrevFullHash(primitives.NewZeroHash())
		newdb.GetHeader().SetPrevKeyMR(primitives.NewZeroHash())
		newdb.GetHeader().SetDBHeight(0)
	}

	newdb.SetDBEntries(make([]interfaces.IDBEntry, 0))

	newdb.AddEntry(primitives.NewHash(constants.ADMIN_CHAINID), primitives.NewZeroHash())
	newdb.AddEntry(primitives.NewHash(constants.EC_CHAINID), primitives.NewZeroHash())
	newdb.AddEntry(primitives.NewHash(constants.FACTOID_CHAINID), primitives.NewZeroHash())

	return newdb
}

func CheckBlockPairIntegrity(block interfaces.IDirectoryBlock, prev interfaces.IDirectoryBlock) error {
	if block == nil {
		return fmt.Errorf("No block specified")
	}

	if prev == nil {
		if block.GetHeader().GetPrevKeyMR().IsZero() == false {
			return fmt.Errorf("Invalid PrevKeyMR")
		}
		if block.GetHeader().GetPrevFullHash().IsZero() == false {
			return fmt.Errorf("Invalid PrevFullHash")
		}
		if block.GetHeader().GetDBHeight() != 0 {
			return fmt.Errorf("Invalid DBHeight")
		}
	} else {
		if block.GetHeader().GetPrevKeyMR().IsSameAs(prev.GetKeyMR()) == false {
			return fmt.Errorf("Invalid PrevKeyMR")
		}
		if block.GetHeader().GetPrevFullHash().IsSameAs(prev.GetFullHash()) == false {
			return fmt.Errorf("Invalid PrevFullHash")
		}
		if block.GetHeader().GetDBHeight() != (prev.GetHeader().GetDBHeight() + 1) {
			return fmt.Errorf("Invalid DBHeight")
		}
	}

	return nil
}

type ExpandedDBlock DirectoryBlock

func (d DirectoryBlock) MarshalJSON() ([]byte, error) {
	d.GetKeyMR()
	d.GetFullHash()

	return json.Marshal(ExpandedDBlock(d))
}
