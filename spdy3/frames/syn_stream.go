package frames

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/SlyMarbo/spdy/common"
)

type SynStreamFrame struct {
	Flags         common.Flags
	StreamID      common.StreamID
	AssocStreamID common.StreamID
	Priority      common.Priority
	Slot          byte
	Header        http.Header
	rawHeader     []byte
}

func (frame *SynStreamFrame) Compress(com common.Compressor) error {
	if frame.rawHeader != nil {
		return nil
	}

	data, err := com.Compress(frame.Header)
	if err != nil {
		return err
	}

	frame.rawHeader = data
	return nil
}

func (frame *SynStreamFrame) Decompress(decom common.Decompressor) error {
	if frame.Header != nil {
		return nil
	}

	header, err := decom.Decompress(frame.rawHeader)
	if err != nil {
		return err
	}

	frame.Header = header
	frame.rawHeader = nil
	return nil
}

func (frame *SynStreamFrame) Name() string {
	return "SYN_STREAM"
}

func (frame *SynStreamFrame) ReadFrom(reader io.Reader) (int64, error) {
	data, err := common.ReadExactly(reader, 18)
	if err != nil {
		return 0, err
	}

	err = controlFrameCommonProcessing(data[:5], SYN_STREAM, common.FLAG_FIN|common.FLAG_UNIDIRECTIONAL)
	if err != nil {
		return 18, err
	}

	// Get and check length.
	length := int(common.BytesToUint24(data[5:8]))
	if length < 10 {
		return 18, common.IncorrectDataLength(length, 10)
	} else if length > common.MAX_FRAME_SIZE-18 {
		return 18, common.FrameTooLarge
	}

	// Read in data.
	header, err := common.ReadExactly(reader, length-10)
	if err != nil {
		return 18, err
	}

	frame.Flags = common.Flags(data[4])
	frame.StreamID = common.StreamID(common.BytesToUint32(data[8:12]))
	frame.AssocStreamID = common.StreamID(common.BytesToUint32(data[12:16]))
	frame.Priority = common.Priority(data[16] >> 5)
	frame.Slot = data[17]
	frame.rawHeader = header

	if !frame.StreamID.Valid() {
		return 18, common.StreamIdTooLarge
	}
	if frame.StreamID.Zero() {
		return 18, common.StreamIdIsZero
	}
	if !frame.AssocStreamID.Valid() {
		return 18, common.StreamIdTooLarge
	}

	return int64(length + 8), nil
}

func (frame *SynStreamFrame) String() string {
	buf := new(bytes.Buffer)
	flags := ""
	if frame.Flags.FIN() {
		flags += " common.FLAG_FIN"
	}
	if frame.Flags.UNIDIRECTIONAL() {
		flags += " FLAG_UNIDIRECTIONAL"
	}
	if flags == "" {
		flags = "[NONE]"
	} else {
		flags = flags[1:]
	}

	buf.WriteString("SYN_STREAM {\n\t")
	buf.WriteString(fmt.Sprintf("Version:              3\n\t"))
	buf.WriteString(fmt.Sprintf("Flags:                %s\n\t", flags))
	buf.WriteString(fmt.Sprintf("Stream ID:            %d\n\t", frame.StreamID))
	buf.WriteString(fmt.Sprintf("Associated Stream ID: %d\n\t", frame.AssocStreamID))
	buf.WriteString(fmt.Sprintf("Priority:             %d\n\t", frame.Priority))
	buf.WriteString(fmt.Sprintf("Slot:                 %d\n\t", frame.Slot))
	buf.WriteString(fmt.Sprintf("Header:               %#v\n}\n", frame.Header))

	return buf.String()
}

func (frame *SynStreamFrame) WriteTo(writer io.Writer) (int64, error) {
	if frame.rawHeader == nil {
		return 0, errors.New("Error: Headers not written.")
	}
	if !frame.StreamID.Valid() {
		return 0, common.StreamIdTooLarge
	}
	if frame.StreamID.Zero() {
		return 0, common.StreamIdIsZero
	}
	if !frame.AssocStreamID.Valid() {
		return 0, common.StreamIdTooLarge
	}

	header := frame.rawHeader
	length := 10 + len(header)
	out := make([]byte, 18)

	out[0] = 128                       // Control bit and Version
	out[1] = 3                         // Version
	out[2] = 0                         // Type
	out[3] = 1                         // Type
	out[4] = byte(frame.Flags)         // Flags
	out[5] = byte(length >> 16)        // Length
	out[6] = byte(length >> 8)         // Length
	out[7] = byte(length)              // Length
	out[8] = frame.StreamID.B1()       // Stream ID
	out[9] = frame.StreamID.B2()       // Stream ID
	out[10] = frame.StreamID.B3()      // Stream ID
	out[11] = frame.StreamID.B4()      // Stream ID
	out[12] = frame.AssocStreamID.B1() // Associated Stream ID
	out[13] = frame.AssocStreamID.B2() // Associated Stream ID
	out[14] = frame.AssocStreamID.B3() // Associated Stream ID
	out[15] = frame.AssocStreamID.B4() // Associated Stream ID
	out[16] = frame.Priority.Byte(3)   // Priority and unused
	out[17] = frame.Slot               // Slot

	err := common.WriteExactly(writer, out)
	if err != nil {
		return 0, err
	}

	err = common.WriteExactly(writer, header)
	if err != nil {
		return 18, err
	}

	return int64(len(header) + 18), nil
}

// SPDY/3.1
type SynStreamFrameV3_1 struct {
	Flags         common.Flags
	StreamID      common.StreamID
	AssocStreamID common.StreamID
	Priority      common.Priority
	Header        http.Header
	rawHeader     []byte
}

func (frame *SynStreamFrameV3_1) Compress(com common.Compressor) error {
	if frame.rawHeader != nil {
		return nil
	}

	data, err := com.Compress(frame.Header)
	if err != nil {
		return err
	}

	frame.rawHeader = data
	return nil
}

func (frame *SynStreamFrameV3_1) Decompress(decom common.Decompressor) error {
	if frame.Header != nil {
		return nil
	}

	header, err := decom.Decompress(frame.rawHeader)
	if err != nil {
		return err
	}

	frame.Header = header
	frame.rawHeader = nil
	return nil
}

func (frame *SynStreamFrameV3_1) Name() string {
	return "SYN_STREAM"
}

func (frame *SynStreamFrameV3_1) ReadFrom(reader io.Reader) (int64, error) {
	data, err := common.ReadExactly(reader, 18)
	if err != nil {
		return 0, err
	}

	err = controlFrameCommonProcessing(data[:5], SYN_STREAM, common.FLAG_FIN|common.FLAG_UNIDIRECTIONAL)
	if err != nil {
		return 18, err
	}

	// Get and check length.
	length := int(common.BytesToUint24(data[5:8]))
	if length < 10 {
		return 18, common.IncorrectDataLength(length, 10)
	} else if length > common.MAX_FRAME_SIZE-18 {
		return 18, common.FrameTooLarge
	}

	// Read in data.
	header, err := common.ReadExactly(reader, length-10)
	if err != nil {
		return 18, err
	}

	frame.Flags = common.Flags(data[4])
	frame.StreamID = common.StreamID(common.BytesToUint32(data[8:12]))
	frame.AssocStreamID = common.StreamID(common.BytesToUint32(data[12:16]))
	frame.Priority = common.Priority(data[16] >> 5)
	frame.rawHeader = header

	if !frame.StreamID.Valid() {
		return 18, common.StreamIdTooLarge
	}
	if frame.StreamID.Zero() {
		return 18, common.StreamIdIsZero
	}
	if !frame.AssocStreamID.Valid() {
		return 18, common.StreamIdTooLarge
	}

	return int64(length + 8), nil
}

func (frame *SynStreamFrameV3_1) String() string {
	buf := new(bytes.Buffer)
	flags := ""
	if frame.Flags.FIN() {
		flags += " common.FLAG_FIN"
	}
	if frame.Flags.UNIDIRECTIONAL() {
		flags += " FLAG_UNIDIRECTIONAL"
	}
	if flags == "" {
		flags = "[NONE]"
	} else {
		flags = flags[1:]
	}

	buf.WriteString("SYN_STREAM {\n\t")
	buf.WriteString(fmt.Sprintf("Version:              3\n\t"))
	buf.WriteString(fmt.Sprintf("Flags:                %s\n\t", flags))
	buf.WriteString(fmt.Sprintf("Stream ID:            %d\n\t", frame.StreamID))
	buf.WriteString(fmt.Sprintf("Associated Stream ID: %d\n\t", frame.AssocStreamID))
	buf.WriteString(fmt.Sprintf("Priority:             %d\n\t", frame.Priority))
	buf.WriteString(fmt.Sprintf("Header:               %#v\n}\n", frame.Header))

	return buf.String()
}

func (frame *SynStreamFrameV3_1) WriteTo(writer io.Writer) (int64, error) {
	if frame.rawHeader == nil {
		return 0, errors.New("Error: Headers not written.")
	}
	if !frame.StreamID.Valid() {
		return 0, common.StreamIdTooLarge
	}
	if frame.StreamID.Zero() {
		return 0, common.StreamIdIsZero
	}
	if !frame.AssocStreamID.Valid() {
		return 0, common.StreamIdTooLarge
	}

	header := frame.rawHeader
	length := 10 + len(header)
	out := make([]byte, 18)

	out[0] = 128                       // Control bit and Version
	out[1] = 3                         // Version
	out[2] = 0                         // Type
	out[3] = 1                         // Type
	out[4] = byte(frame.Flags)         // Flags
	out[5] = byte(length >> 16)        // Length
	out[6] = byte(length >> 8)         // Length
	out[7] = byte(length)              // Length
	out[8] = frame.StreamID.B1()       // Stream ID
	out[9] = frame.StreamID.B2()       // Stream ID
	out[10] = frame.StreamID.B3()      // Stream ID
	out[11] = frame.StreamID.B4()      // Stream ID
	out[12] = frame.AssocStreamID.B1() // Associated Stream ID
	out[13] = frame.AssocStreamID.B2() // Associated Stream ID
	out[14] = frame.AssocStreamID.B3() // Associated Stream ID
	out[15] = frame.AssocStreamID.B4() // Associated Stream ID
	out[16] = frame.Priority.Byte(3)   // Priority and unused
	out[17] = 0                        // Reserved

	err := common.WriteExactly(writer, out)
	if err != nil {
		return 0, err
	}

	err = common.WriteExactly(writer, header)
	if err != nil {
		return 18, err
	}

	return int64(len(header) + 18), nil
}
