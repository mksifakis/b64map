## b64map

Maps the given program over the input the input. Standard input and output are
expected to be base 64 encoded, one document or record per line. The program
is run on each line of input and then re-encoded.

For example:

    $ cat > test1 <<EOF
    Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
    Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
    Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
    Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
    EOF
    $ cat > test2 <<EOF
    Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.
    Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.
    Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem.
    EOF
    $ < test1 base64 -w 0 > test; echo >> test
    $ < test2 base64 -w 0 >> test; echo >> test
    $ wc -l test
        2 test

    $ < test b64filter cat > test.cat
    2020/02/24 12:15:29 b64map.go:180: processed 2 documents

The program can be installed with

    go get github.com/paracrawl/b64map
    
