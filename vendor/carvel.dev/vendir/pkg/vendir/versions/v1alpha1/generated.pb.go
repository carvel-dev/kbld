// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: carvel.dev/vendir/pkg/vendir/versions/v1alpha1/generated.proto

package v1alpha1

import (
	fmt "fmt"

	io "io"
	math "math"
	math_bits "math/bits"
	reflect "reflect"
	strings "strings"

	proto "github.com/gogo/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

func (m *VersionSelection) Reset()      { *m = VersionSelection{} }
func (*VersionSelection) ProtoMessage() {}
func (*VersionSelection) Descriptor() ([]byte, []int) {
	return fileDescriptor_f7fa722d77d11bd9, []int{0}
}
func (m *VersionSelection) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *VersionSelection) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *VersionSelection) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VersionSelection.Merge(m, src)
}
func (m *VersionSelection) XXX_Size() int {
	return m.Size()
}
func (m *VersionSelection) XXX_DiscardUnknown() {
	xxx_messageInfo_VersionSelection.DiscardUnknown(m)
}

var xxx_messageInfo_VersionSelection proto.InternalMessageInfo

func (m *VersionSelectionSemver) Reset()      { *m = VersionSelectionSemver{} }
func (*VersionSelectionSemver) ProtoMessage() {}
func (*VersionSelectionSemver) Descriptor() ([]byte, []int) {
	return fileDescriptor_f7fa722d77d11bd9, []int{1}
}
func (m *VersionSelectionSemver) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *VersionSelectionSemver) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *VersionSelectionSemver) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VersionSelectionSemver.Merge(m, src)
}
func (m *VersionSelectionSemver) XXX_Size() int {
	return m.Size()
}
func (m *VersionSelectionSemver) XXX_DiscardUnknown() {
	xxx_messageInfo_VersionSelectionSemver.DiscardUnknown(m)
}

var xxx_messageInfo_VersionSelectionSemver proto.InternalMessageInfo

func (m *VersionSelectionSemverPrereleases) Reset()      { *m = VersionSelectionSemverPrereleases{} }
func (*VersionSelectionSemverPrereleases) ProtoMessage() {}
func (*VersionSelectionSemverPrereleases) Descriptor() ([]byte, []int) {
	return fileDescriptor_f7fa722d77d11bd9, []int{2}
}
func (m *VersionSelectionSemverPrereleases) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *VersionSelectionSemverPrereleases) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *VersionSelectionSemverPrereleases) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VersionSelectionSemverPrereleases.Merge(m, src)
}
func (m *VersionSelectionSemverPrereleases) XXX_Size() int {
	return m.Size()
}
func (m *VersionSelectionSemverPrereleases) XXX_DiscardUnknown() {
	xxx_messageInfo_VersionSelectionSemverPrereleases.DiscardUnknown(m)
}

var xxx_messageInfo_VersionSelectionSemverPrereleases proto.InternalMessageInfo

func init() {
	proto.RegisterType((*VersionSelection)(nil), "carvel.dev.vendir.pkg.vendir.versions.v1alpha1.VersionSelection")
	proto.RegisterType((*VersionSelectionSemver)(nil), "carvel.dev.vendir.pkg.vendir.versions.v1alpha1.VersionSelectionSemver")
	proto.RegisterType((*VersionSelectionSemverPrereleases)(nil), "carvel.dev.vendir.pkg.vendir.versions.v1alpha1.VersionSelectionSemverPrereleases")
}

func init() {
	proto.RegisterFile("carvel.dev/vendir/pkg/vendir/versions/v1alpha1/generated.proto", fileDescriptor_f7fa722d77d11bd9)
}

var fileDescriptor_f7fa722d77d11bd9 = []byte{
	// 333 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x8f, 0x31, 0x4b, 0xfb, 0x40,
	0x18, 0xc6, 0x73, 0xff, 0x3f, 0x14, 0x7b, 0x19, 0x94, 0x08, 0x52, 0x1c, 0xae, 0xda, 0xa9, 0x8b,
	0x17, 0x2a, 0xb8, 0x3a, 0x44, 0x10, 0xdc, 0x34, 0x85, 0x0e, 0x6e, 0x69, 0xf2, 0x36, 0x3d, 0x9b,
	0xde, 0x85, 0xbb, 0x6b, 0x46, 0x71, 0xf0, 0x03, 0xf8, 0xb1, 0x3a, 0x76, 0x92, 0x4e, 0xc5, 0x9e,
	0x5f, 0x44, 0x7a, 0x31, 0x24, 0x88, 0x20, 0x82, 0xdb, 0x73, 0xef, 0xbd, 0xcf, 0xef, 0x7d, 0x1e,
	0x7c, 0x19, 0x47, 0xb2, 0x80, 0x8c, 0x26, 0x50, 0xf8, 0x05, 0xf0, 0x84, 0x49, 0x3f, 0x9f, 0xa5,
	0x95, 0x2c, 0x40, 0x2a, 0x26, 0xb8, 0xf2, 0x8b, 0x41, 0x94, 0xe5, 0xd3, 0x68, 0xe0, 0xa7, 0xc0,
	0x41, 0x46, 0x1a, 0x12, 0x9a, 0x4b, 0xa1, 0x85, 0x47, 0x6b, 0x3f, 0x2d, 0x4d, 0x34, 0x9f, 0xa5,
	0x95, 0xac, 0xfc, 0xb4, 0xf2, 0x1f, 0x9f, 0xa5, 0x4c, 0x4f, 0x17, 0x63, 0x1a, 0x8b, 0xb9, 0x9f,
	0x8a, 0x54, 0xf8, 0x16, 0x33, 0x5e, 0x4c, 0xec, 0xcb, 0x3e, 0xac, 0x2a, 0xf1, 0xbd, 0x47, 0x7c,
	0x30, 0x2a, 0x19, 0x43, 0xc8, 0x20, 0xd6, 0x4c, 0x70, 0xef, 0x01, 0xb7, 0x14, 0xcc, 0x0b, 0x90,
	0x1d, 0x74, 0x82, 0xfa, 0xee, 0xf9, 0xf5, 0x2f, 0x33, 0xd0, 0xaf, 0xc4, 0xa1, 0xa5, 0x05, 0xd8,
	0x6c, 0xba, 0xad, 0x52, 0x87, 0x9f, 0x17, 0x7a, 0xaf, 0x08, 0x1f, 0x7d, 0xbf, 0xee, 0x5d, 0x60,
	0x37, 0x16, 0x5c, 0x69, 0x19, 0x31, 0xae, 0x95, 0xcd, 0xd2, 0x0e, 0x0e, 0x97, 0x9b, 0xae, 0x63,
	0x36, 0x5d, 0xf7, 0xaa, 0xfe, 0x0a, 0x9b, 0x7b, 0xde, 0x33, 0xc2, 0x6e, 0x2e, 0x41, 0x42, 0x06,
	0x91, 0x02, 0xd5, 0xf9, 0x67, 0x3b, 0xdc, 0xfd, 0x4d, 0x87, 0xdb, 0x1a, 0x1c, 0xec, 0xef, 0x62,
	0x34, 0x06, 0x61, 0xf3, 0x6c, 0x6f, 0x84, 0x4f, 0x7f, 0x44, 0x78, 0x03, 0xec, 0xb2, 0x04, 0xb8,
	0x66, 0x13, 0x06, 0x72, 0x57, 0xf1, 0x7f, 0xbf, 0x5d, 0x72, 0x6f, 0xea, 0x71, 0xd8, 0xdc, 0x09,
	0xe8, 0x72, 0x4b, 0x9c, 0xd5, 0x96, 0x38, 0xeb, 0x2d, 0x71, 0x9e, 0x0c, 0x41, 0x4b, 0x43, 0xd0,
	0xca, 0x10, 0xb4, 0x36, 0x04, 0xbd, 0x19, 0x82, 0x5e, 0xde, 0x89, 0x73, 0xbf, 0x57, 0xf5, 0xf8,
	0x08, 0x00, 0x00, 0xff, 0xff, 0xff, 0x5f, 0x50, 0x3c, 0x80, 0x02, 0x00, 0x00,
}

func (m *VersionSelection) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *VersionSelection) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *VersionSelection) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Semver != nil {
		{
			size, err := m.Semver.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenerated(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *VersionSelectionSemver) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *VersionSelectionSemver) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *VersionSelectionSemver) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Prereleases != nil {
		{
			size, err := m.Prereleases.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenerated(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	i -= len(m.Constraints)
	copy(dAtA[i:], m.Constraints)
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.Constraints)))
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *VersionSelectionSemverPrereleases) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *VersionSelectionSemverPrereleases) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *VersionSelectionSemverPrereleases) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Identifiers) > 0 {
		for iNdEx := len(m.Identifiers) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Identifiers[iNdEx])
			copy(dAtA[i:], m.Identifiers[iNdEx])
			i = encodeVarintGenerated(dAtA, i, uint64(len(m.Identifiers[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenerated(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenerated(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *VersionSelection) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Semver != nil {
		l = m.Semver.Size()
		n += 1 + l + sovGenerated(uint64(l))
	}
	return n
}

func (m *VersionSelectionSemver) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Constraints)
	n += 1 + l + sovGenerated(uint64(l))
	if m.Prereleases != nil {
		l = m.Prereleases.Size()
		n += 1 + l + sovGenerated(uint64(l))
	}
	return n
}

func (m *VersionSelectionSemverPrereleases) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Identifiers) > 0 {
		for _, s := range m.Identifiers {
			l = len(s)
			n += 1 + l + sovGenerated(uint64(l))
		}
	}
	return n
}

func sovGenerated(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenerated(x uint64) (n int) {
	return sovGenerated(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *VersionSelection) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&VersionSelection{`,
		`Semver:` + strings.Replace(this.Semver.String(), "VersionSelectionSemver", "VersionSelectionSemver", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *VersionSelectionSemver) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&VersionSelectionSemver{`,
		`Constraints:` + fmt.Sprintf("%v", this.Constraints) + `,`,
		`Prereleases:` + strings.Replace(this.Prereleases.String(), "VersionSelectionSemverPrereleases", "VersionSelectionSemverPrereleases", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *VersionSelectionSemverPrereleases) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&VersionSelectionSemverPrereleases{`,
		`Identifiers:` + fmt.Sprintf("%v", this.Identifiers) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringGenerated(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *VersionSelection) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: VersionSelection: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: VersionSelection: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Semver", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Semver == nil {
				m.Semver = &VersionSelectionSemver{}
			}
			if err := m.Semver.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenerated
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *VersionSelectionSemver) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: VersionSelectionSemver: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: VersionSelectionSemver: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Constraints", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Constraints = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Prereleases", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Prereleases == nil {
				m.Prereleases = &VersionSelectionSemverPrereleases{}
			}
			if err := m.Prereleases.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenerated
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *VersionSelectionSemverPrereleases) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: VersionSelectionSemverPrereleases: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: VersionSelectionSemverPrereleases: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Identifiers", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Identifiers = append(m.Identifiers, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenerated
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGenerated(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthGenerated
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenerated
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenerated
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenerated        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenerated          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenerated = fmt.Errorf("proto: unexpected end of group")
)