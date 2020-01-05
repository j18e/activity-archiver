package main

import (
	"testing"
	"time"
)

var data = `: 1573742005:0;git branch -a
: 1573742199:0;git rm --cached myapp/src/images/photos/shutterstock_80395672.jpg
: 1573742243:0;git reset --hard
: 1573742245:0;git rm --cached myapp/src/images/photos/shutterstock_80395672.jpg myapp/src/images/photos/shutterstock_98595776.jpg myapp/src/images/photos/shutterstock_197734370.eps myapp/src/images/photos/shutterstock_568586704.jpg myapp/src/images/photos/shutterstock_178882499.jpg myapp/src/images/photos/shutterstock_516126295.eps myapp/src/images/photos/shutterstock_739989259.jpg myapp/src/images/photos/shutterstock_794240764.jpg myapp/src/images/photos/shutterstock_621875594.jpg myapp/src/images/photos/shutterstock_743692444.jpg myapp/src/images/photos/shutterstock_585257993.jpg myapp/src/images/photos/shutterstock_265381130.jpg myapp/src/images/photos/shutterstock_611857931.jpg myapp/src/images/photos/shutterstock_135442784.jpg myapp/src/images/photos/shutterstock_154142756.jpg myapp/src/images/photos/shutterstock_192237287.jpg myapp/src/images/photos/shutterstock_397902607.jpg myapp/src/images/photos/shutterstock_1057495028.jpg
: 1573742261:0;ga myapp/myapp-web.csproj
: 1573742272:0;gcm "remove large, unused photos"
: 1573742282:0;git clean -f -d
: 1573742533:0;git filter-branch --force --index-filter \\
'git rm --cached --ignore-unmatch myapp/src/images/photos/shutterstock_*' \\
--prune-empty --tag-name-filter cat -- --all\
\
\
git for-each-ref --format='delete %(refname)' refs/original | git update-ref --stdin\
\
git reflog expire --expire=now --all\

: 1573742761:0;git reflog expire --expire=now --all
: 1573742772:0;git gc --prune=now --aggressive
: 1573742821:0;sed -i '/shutterstock_/d' myapp/myapp-web.csproj\

: 1573742824:0;gs
: 1573742829:0;sed -i '/shutterstock_/d' myapp/myapp-web.csproj
: 1573742889:0;git remote add origin ssh://git@github.com/myuser/web-frontend.git
: 1573742894:0;gp
: 1573742913:0;l tmp
: 1573742915:0;cd sandbox`

func TestParseLine(t *testing.T) {
	expTS := time.Unix(1573742913, 0)
	expCmd := "ls tmp"

	line := ": 1573742913:0;ls tmp"

	res, err := parseLine(line)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	if res.Timestamp != expTS {
		t.Errorf("expected timestamp %v, got %v", expTS, res.Timestamp)
	}
	if res.Command != expCmd {
		t.Errorf("expected cmd %s, got %s", res.Command, expCmd)
	}
}
