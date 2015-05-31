###########################################################
## generate proto payload struct 
##
@include "header.awk"
BEGIN { RS = "==="; FS ="\n" 
print ""
print "import \"misc/packet\"\n"
}
{

	pack_code = ""
	for (i=1;i<=NF;i++)
	{
		if ($i ~ /[A-Za-z_]+=/) {
			name = substr($i,1, match($i,/=/)-1)
			print "type " name " struct {"
			typeok = "true"
		} else {
			if ($i!="" && typeok) {	
				print(_field($i))
				pack_code = pack_code _writer($i)
			}
		}
	}

	if (typeok) {
		print "}\n"
		print "func (p " name ") Pack(w *packet.Packet) {"
		print pack_code
		print "}"
	}

	typeok=false
}
END { }

function _field(line) {
	split(line, a, " ")

	if (a[2] in TYPES) {
		return "\tF_"a[1] " " TYPES[a[2]]
	} else if (a[2] == "array") {
		if (a[3] in TYPES) {
			return "\tF_"a[1]" []" TYPES[a[3]]
		} else {
			return "\tF_"a[1]" []" a[3]
		}
	} else {
		return "\tF_"a[1]" " a[2]
	}
}

function _writer(line) {
	split(line, a, " ")

	if (a[2] in WRITERS) {
		return "w." WRITERS[a[2]] "(p.F_" a[1] ")\n"
	} else if (a[2] == "array") {
		ret = "w.WriteU16(uint16(len(p.F_" a[1] ")))\n"
		if (a[3] == "byte") {
			ret = ret "w.WriteRawBytes(p.F_"a[1]")\n"
			return ret
		} else if (a[3] in TYPES) {
			ret = ret "for k:=range p.F_" a[1] "{\n"
				ret = ret "w." WRITERS[a[3]] "(p.F_" a[1] "[k])\n"
			return ret "}\n"
		} else {
			ret = ret "for k:=range p.F_" a[1] "{\n"
				ret = ret "p.F_"a[1]"[k].Pack(w)\n"
			return ret "}\n"
		}
	} else {
		return "p.F_" a[1] ".Pack(w)\n"
	}
}
