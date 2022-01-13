package infoloader

import "testing"

func TestXsplit(t *testing.T) {
	vs := XSplit("engine sip launch", 2)
	if vs[1] != "sip launch" {
		t.Fail()
	}

	vs = XSplit("    enge,sdfwette, 32423  low(>>>   <<, ", 3)
	if len(vs) != 3 || vs[0] != "enge,sdfwette," || vs[1] != "32423" ||
		vs[2] != "low(>>>   <<," {
		t.Logf("%+v", vs)
		t.Fail()
	}

	vs = XSplit("    enge,sdfwette, 32423  low(>>>   <<, ", 4)
	if len(vs) != 4 || vs[0] != "enge,sdfwette," || vs[1] != "32423" ||
		vs[2] != "low(>>>" || vs[3] != "<<," {
		t.Logf("%+v", vs)
		t.Fail()
	}

	vs = XSplit("a b c d ", 4)
	if len(vs) != 4 || vs[0] != "a" || vs[1] != "b" || vs[2] != "c" || vs[3] != "d" {
		t.Fail()
	}

	vs = XSplit("  ,  &%#*#a, b, cde stringhello", 4)
	if len(vs) != 4 || vs[3] != "cde stringhello" ||
		vs[2] != "b," ||
		vs[0] != "," || vs[1] != "&%#*#a," {
		t.Fail()
	}

	vs = XSplit("  ,  &%#*#a  ", 10)
	if len(vs) != 2 || vs[1] != "&%#*#a" {
		t.Fail()
	}

	vs = XSplit("a b c d ", 20)
	if len(vs) != 4 || vs[0] != "a" || vs[1] != "b" || vs[2] != "c" || vs[3] != "d" {
		t.Fail()
	}

	vs = XSplit("a b c d e", 20)
	if len(vs) != 5 {
		t.Fail()
	}
}
