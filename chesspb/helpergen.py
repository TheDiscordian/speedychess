pb = open('../chesspb/chesspb.pb.go').read().split('\n')
pbtypes = """package chesspb

import "google.golang.org/protobuf/proto"

const (
    _ = iota - 1
"""
types = []
prefix = "\t(*"
for i in pb:
    if len(i) > len(prefix) and i[:len(prefix)] == prefix:
        types.append(i[len(prefix):len(prefix)+i[len(prefix):].index(")")])
        pbtypes += "    %sMsg\n" % types[-1]

pbtypes += """)

func NewMsg(t byte) (proto.Message) {
    switch t {
"""
for i in types:
    pbtypes += """    case %sMsg:
        return new(%s)
""" % (i, i)
pbtypes += """    }
    return nil
}
"""
open("../chesspb/pbgen.go", "w").write(pbtypes)
