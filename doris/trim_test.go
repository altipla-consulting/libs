package doris

import (
	"testing"

	"github.com/stretchr/testify/require"

	pb "libs.altipla.consulting/hosting/testdata"
)

func TestTrimStrings(t *testing.T) {
	in := &pb.TrimTest{
		Direct: "  direct\n\t",
		Sub: &pb.Message{
			Child:         "  child",
			RepeatedChild: []string{"  repeated child1", "repeated child2  "},
		},
		Map: map[string]string{
			"  key": "  value",
		},
		List: []string{"  item"},
		Mapsub: map[string]*pb.Message{
			"  key": {
				Child: "  mapsub child",
			},
		},
		Oneof1: &pb.TrimTest_OneofDirect1{
			OneofDirect1: "  oneof_direct1",
		},
		Oneof2: &pb.TrimTest_OneofSub2{
			OneofSub2: &pb.Message{
				Child: "  oneof_child",
			},
		},
		RepeatedMessage: []*pb.Message{
			{Child: "  repeated1"},
			{Child: "repeated2  "},
		},
		Enum: pb.TrimTest_TEST_FOO,
	}
	out := trimMessage(in.ProtoReflect()).Interface().(*pb.TrimTest)
	require.Equal(t, out.Direct, "direct")
	require.Equal(t, out.Sub.Child, "child")
	require.Equal(t, out.Sub.RepeatedChild, []string{"repeated child1", "repeated child2"})
	require.Equal(t, out.Map, map[string]string{"key": "value"})
	require.Nil(t, out.Empty)
	require.Equal(t, out.List, []string{"item"})
	require.Contains(t, out.Mapsub, "key")
	require.Equal(t, out.Mapsub["key"].Child, "mapsub child")
	require.Len(t, out.Mapsub, 1)
	require.Equal(t, out.GetOneofDirect1(), "oneof_direct1")
	require.Equal(t, out.GetOneofSub2().Child, "oneof_child")
	require.Len(t, out.RepeatedMessage, 2)
	require.Equal(t, out.RepeatedMessage[0].Child, "repeated1")
	require.Equal(t, out.RepeatedMessage[1].Child, "repeated2")
	require.Equal(t, out.Enum, pb.TrimTest_TEST_FOO)
}
