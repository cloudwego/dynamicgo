package generic

import (
	"github.com/cloudwego/dynamicgo/proto"
)

var (
	PathInnerBase              = NewPathFieldName("InnerBase2")
	PathNotExist               = []Path{NewPathFieldName("NotExist")}
	PathExampleBool            = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(1))}
	PathExampleUint32          = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(2))}
	PathExampleUint64          = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(3))}
	PathExampleInt32           = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(4))}
	PathExampleInt64           = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(5))}
	PathExampleDouble          = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(6))}
	PathExampleString          = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(7))}
	PathExampleListInt32       = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(8))}
	PathExampleMapStringString = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(9))}
	PathExampleSetInt32_       = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(10))}
	PathExampleFoo             = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(11))}
	PathExampleMapInt32String  = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(12))}
	PathExampleBinary          = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(13))}
	PathExampleBase            = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(255))}
)