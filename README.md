Instructions
------------

Run web server

    cd client
    python -m SimpleHTTPServer 8000

Run application server

    export ALLOWED_ORIGIN=http://127.0.0.1:8000
    go run server.go

Visit [index.html](http://127.0.0.1:8000) in your browser.


Post-mortem
-----------

For debugging and testing I used Wireshark, CURL, python scripts, Firebug.
