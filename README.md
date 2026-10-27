# polls2026
US Senate Polls for 2026
<br>
Just having fun with sampling various political prognosticator data.
Take everything with a YUGE grain of salt.

### Installation

git clone https://github.com/texadactyl/polls2026
<br>
cd polls2026
<br>
go install .
<br>
(or, go build -o `<somewhere in the path>` .)

### Verification

go test -v .
<br>
polls2026 -l
<br>
polls2026 -r 1 AK
<br>
polls2026 -r 2 0

### Documentation

Read the GoLang code, starting with main.go.
