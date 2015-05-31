###########################################################
## generate protocol packet reader
##
@include "header.awk"
BEGIN { RS = ""; FS ="\n"}
{
	for (i=1;i<=NF;i++)
	{
		if ($i ~ /^#.*/ || $i ~ /^===/) {
			continue
		}
		split($i, a, " ")
		if (a[1] ~ /[A-Za-z_]+=/) {
			name = substr(a[1],1, match(a[1],/=/)-1)
			print "func PKT_"name"(reader *packet.Packet)(tbl "name", err error){"
			typeok = "true"
		} else if (a[2] ==  "array") {
			if (a[3] == "byte") { 		## bytes
				print "\ttbl.F_"a[1]", err = reader.ReadBytes()"
				print "\tcheckErr(err)\n"
			} else if (a[3] in READERS) {	## primitives
				print "\t{"
				print "\tnarr,err := reader.ReadU16()"
				print "\tcheckErr(err)\n"
				print "\tfor i:=0;i<int(narr);i++ {"
				print "\t\tv, err := reader."READERS[a[3]]"()"
				print "\t\ttbl.F_"a[1]" = append(tbl.F_"a[1]", v)"
				print "\t\tcheckErr(err)\n"
				print "\t}\n"
				print "\t}\n"
			} else {	## struct
				print "\t{"
				print "\tnarr,err := reader.ReadU16()"
				print "\tcheckErr(err)\n"
				print "\ttbl.F_"a[1]"=make([]"a[3]",narr)"
				print "\tfor i:=0;i<int(narr);i++ {"
				print "\t\ttbl.F_"a[1]"[i], err = PKT_"a[3]"(reader)"
				print "\t\tcheckErr(err)\n"
				print "\t}\n"
				print "\t}\n"
			}
		}
		else if (!(a[2] in READERS)) {
			print "\t\ttbl.F_"a[1]", err = PKT_"a[2]"(reader)"
			print "\tcheckErr(err)\n"
		}
		else {
			print "\ttbl.F_"a[1]",err = reader." READERS[a[2]] "()"
			print "\tcheckErr(err)\n"
		}
	}

	if (typeok) {
		print "\treturn"
		print "}\n"
	}

	typeok=false
}
END { }
